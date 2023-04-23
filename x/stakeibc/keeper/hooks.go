package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v9/utils"
	epochstypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// Update the stakeibc epoch tracker
	epochNumber, err := k.UpdateEpochTracker(ctx, epochInfo)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to update epoch tracker, err: %s", err.Error()))
		return
	}

	// Day Epoch - Process Unbondings
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		// Initiate unbondings from any hostZone where it's appropriate
		k.InitiateAllHostZoneUnbondings(ctx, epochNumber)
		// Check previous epochs to see if unbondings finished, and sweep the tokens if so
		k.SweepAllUnbondedTokens(ctx)
		// Cleanup any records that are no longer needed
		k.CleanupEpochUnbondingRecords(ctx, epochNumber)
		// Create an empty unbonding record for this epoch
		k.CreateEpochUnbondingRecord(ctx, epochNumber)
	}

	// Stride Epoch - Process Deposits and Delegations
	if epochInfo.Identifier == epochstypes.STRIDE_EPOCH {
		// Get cadence intervals
		redemptionRateInterval := k.GetParam(ctx, types.KeyRedemptionRateInterval)
		depositInterval := k.GetParam(ctx, types.KeyDepositInterval)
		delegationInterval := k.GetParam(ctx, types.KeyDelegateInterval)
		reinvestInterval := k.GetParam(ctx, types.KeyReinvestInterval)

		// Create a new deposit record for each host zone and the grab all deposit records
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)
		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)

		// TODO: move this to an external function that anyone can call, so that we don't have to call it every epoch
		k.SetWithdrawalAddress(ctx)

		// Update the redemption rate
		if epochNumber%redemptionRateInterval == 0 {
			k.UpdateRedemptionRates(ctx, depositRecords)
		}

		// Transfer deposited funds from the controller account to the delegation account on the host zone
		if epochNumber%depositInterval == 0 {
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}

		// Delegate tokens from the delegation account
		if epochNumber%delegationInterval == 0 {
			k.StakeExistingDepositsOnHostZones(ctx, epochNumber, depositRecords)
		}

		// Reinvest staking rewards
		if epochNumber%reinvestInterval == 0 { // allow a few blocks from UpdateUndelegatedBal to avoid conflicts
			k.ReinvestRewards(ctx)
		}
	}
	if epochInfo.Identifier == epochstypes.MINT_EPOCH {
		k.AllocateHostZoneReward(ctx)
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}

// Update the epoch information in the stakeibc epoch tracker
func (k Keeper) UpdateEpochTracker(ctx sdk.Context, epochInfo epochstypes.EpochInfo) (epochNumber uint64, err error) {
	epochNumber, err = cast.ToUint64E(epochInfo.CurrentEpoch)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch number to uint64: %v", err))
		return 0, err
	}
	epochDurationNano, err := cast.ToUint64E(epochInfo.Duration.Nanoseconds())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return 0, err
	}
	nextEpochStartTime, err := cast.ToUint64E(epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration).UnixNano())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return 0, err
	}
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochInfo.Identifier,
		EpochNumber:        epochNumber,
		Duration:           epochDurationNano,
		NextEpochStartTime: nextEpochStartTime,
	}
	k.SetEpochTracker(ctx, epochTracker)

	return epochNumber, nil
}

// Set the withdrawal account address for each host zone
func (k Keeper) SetWithdrawalAddress(ctx sdk.Context) {
	k.Logger(ctx).Info("Setting Withdrawal Addresses...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		err := k.SetWithdrawalAddressOnHost(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to set withdrawal address on %s, err: %s", hostZone.ChainId, err))
		}
	}
}

// Updates the redemption rate for each host zone
// At a high level, the redemption rate is equal to the amount of native tokens locked divided by the stTokens in existence.
// The equation is broken down further into the following sub-components:
//     Native Tokens Locked:
//       1. Deposit Account Balance: native tokens deposited from liquid stakes, that are still living on Stride
//       2. Undelegated Balance:     native tokens that have been transferred to the host zone, but have not been delegated yet
//       3. Tokenized Delegations:   Delegations inherent in LSM Tokens that have not yet been converted to native stake
//       4. Unbalanced Delegations:  Delegations that have been converted from LSM Tokens to native stake, but have not yet been balanced
//       5. Balanced Delegations:    Delegations either from native tokens, or LSM Tokens that have undergone rebalancing
//    StToken Amount:
//       1. Total Supply of the stToken
//
//  Redemption Rate =
//  (Deposit Account Balance + Undelegated Balance + Tokenized Delegation + Unbalanced Delegation + Balanced Delegation) / (stToken Supply)
func (k Keeper) UpdateRedemptionRates(ctx sdk.Context, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Updating Redemption Rates...")

	// Update the redemption rate for each host zone
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {

		// Gather redemption rate components
		stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)).Amount
		if stSupply.IsZero() {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
				"No st%s in circulation - redemption rate is unchanged", hostZone.HostDenom))
			continue
		}

		depositAccountBalance := k.GetDepositAccountBalance(hostZone.ChainId, depositRecords)
		undelegatedBalance := k.GetUndelegatedBalance(hostZone.ChainId, depositRecords)
		tokenizedDelegation := k.GetTotalTokenizedDelegations(ctx, hostZone.ChainId)
		balancedDelegation := sdk.NewDecFromInt(hostZone.TotalBalancedDelegations)
		unbalancedDelegation := sdk.NewDecFromInt(hostZone.TotalUnbalancedDelegations)

		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Redemption Rate Components - "+
			"Deposit Account Balance: %v, Undelegated Balance: %v, LSM Delegated Balance: %v"+
			"Native Balanced Delegations: %v, Native Unbalanced Delegations: %v, stToken Supply: %v",
			depositAccountBalance, undelegatedBalance, tokenizedDelegation,
			balancedDelegation, unbalancedDelegation, stSupply))

		// Calculate the redemption rate
		nativeTokensLocked := depositAccountBalance.
			Add(undelegatedBalance).
			Add(tokenizedDelegation).
			Add(balancedDelegation).
			Add(unbalancedDelegation)
		redemptionRate := nativeTokensLocked.Quo(sdk.NewDecFromInt(stSupply))

		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"New Redemption Rate: %v (vs Prev Rate: %v)", redemptionRate, hostZone.RedemptionRate))

		// Update the host zone
		hostZone.LastRedemptionRate = hostZone.RedemptionRate
		hostZone.RedemptionRate = redemptionRate
		k.SetHostZone(ctx, hostZone)
	}
}

// Determine the deposit account balance, representing native tokens that have been deposited
// from liquid stakes, but have not yet been transferred to the host
func (k Keeper) GetDepositAccountBalance(chainId string, depositRecords []recordstypes.DepositRecord) sdk.Dec {
	// sum on deposit records with status TRANSFER_QUEUE or TRANSFER_IN_PROGRESS
	totalAmount := sdkmath.ZeroInt()
	for _, depositRecord := range depositRecords {
		transferStatus := (depositRecord.Status == recordstypes.DepositRecord_TRANSFER_QUEUE ||
			depositRecord.Status == recordstypes.DepositRecord_TRANSFER_IN_PROGRESS)

		if depositRecord.HostZoneId == chainId && transferStatus {
			totalAmount = totalAmount.Add(depositRecord.Amount)
		}
	}

	return sdk.NewDecFromInt(totalAmount)
}

// Determine the undelegated balance from the deposit records queued for staking
func (k Keeper) GetUndelegatedBalance(chainId string, depositRecords []recordstypes.DepositRecord) sdk.Dec {
	// sum on deposit records with status DELEGATION_QUEUE or DELEGATION_IN_PROGRESS
	totalAmount := sdkmath.ZeroInt()
	for _, depositRecord := range depositRecords {
		delegationStatus := (depositRecord.Status == recordstypes.DepositRecord_DELEGATION_QUEUE ||
			depositRecord.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS)

		if depositRecord.HostZoneId == chainId && delegationStatus {
			totalAmount = totalAmount.Add(depositRecord.Amount)
		}
	}

	return sdk.NewDecFromInt(totalAmount)
}

// Returns the total delegated balance that's stored in LSM tokens
// These are identified by any status besides "DEPOSIT_PENDING"
func (k Keeper) GetTotalTokenizedDelegations(ctx sdk.Context, chainId string) sdk.Dec {
	total := sdkmath.ZeroInt()
	for _, deposit := range k.RecordsKeeper.GetLSMDepositsForHostZone(ctx, chainId) {
		if deposit.Status != recordstypes.LSMTokenDeposit_DEPOSIT_PENDING {
			total = total.Add(deposit.Amount)
		}
	}

	return sdk.NewDecFromInt(total)
}

func (k Keeper) ReinvestRewards(ctx sdk.Context) {
	k.Logger(ctx).Info("Reinvesting tokens...")

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		// only process host zones once withdrawal accounts are registered
		if hostZone.WithdrawalIcaAddress == "" {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Withdrawal account not registered for host zone"))
			continue
		}

		// read clock time on host zone
		blockTime, err := k.GetLightClientTimeSafely(ctx, hostZone.ConnectionId)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not find blockTime for host zone %s, err: %s", hostZone.ConnectionId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "BlockTime for host zone: %d", blockTime))

		err = k.UpdateWithdrawalBalance(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error updating withdrawal balance for host zone %s: %s", hostZone.ConnectionId, err.Error()))
			continue
		}
	}
}
