package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// TODO [TEST-127]: ensure all timeouts are less than the epoch length
// TODO [TEST-126]: add events from event manager, e.g.
// ctx.EventManager().EmitEvents(sdk.Events{
// 	sdk.NewEvent(
// 		sdk.EventTypeMessage,
// 		sdk.NewAttribute("hostZone", zoneInfo.ChainId),
// 		sdk.NewAttribute("newAmountStaked", balance.String()),
// 	),
// })

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// every epoch
	epochIdentifier := epochInfo.Identifier
	epochNumber, err := cast.ToUint64E(epochInfo.CurrentEpoch)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch number to uint64: %v", err))
		return
	}

	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s %d", epochIdentifier, epochNumber))
	k.Logger(ctx).Info(fmt.Sprintf("Epoch start time %d", epochInfo.GetCurrentEpochStartTime().UnixNano()))

	ns, err := cast.ToUint64E(epochInfo.GetDuration().Nanoseconds())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return
	}
	nextEpochStartTime, err := cast.ToUint64E(epochInfo.GetCurrentEpochStartTime().Add(epochInfo.GetDuration()).UnixNano())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return
	}
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        epochNumber,
		Duration:           ns,
		NextEpochStartTime: nextEpochStartTime,
	}
	// deposit records *must* exist for this epoch
	k.Logger(ctx).Info(fmt.Sprintf("Setting epochTracker %v", epochTracker))
	k.SetEpochTracker(ctx, epochTracker)

	// process redemption records
	if epochIdentifier == epochstypes.DAY_EPOCH {
		// here, we process everything we need to for redemptions
		k.Logger(ctx).Info(utils.LogHeader("DAY EPOCH %d", epochNumber))
		// first we initiate unbondings from any hostZone where it's appropriate
		k.Logger(ctx).Info("InitiateAllHostZoneUnbondings")
		k.InitiateAllHostZoneUnbondings(ctx, epochNumber)
		// then we check previous epochs to see if unbondings finished, and sweep the tokens if so
		k.Logger(ctx).Info("SweepAllUnbondedTokens")
		k.SweepAllUnbondedTokens(ctx)
		// then we cleanup any records that are no longer needed
		k.Logger(ctx).Info("CleanupEpochUnbondingRecords")
		k.CleanupEpochUnbondingRecords(ctx, epochNumber)
		// lastly we create an empty unbonding record for this epoch
		k.Logger(ctx).Info("CreateEpochUnbondingRecord")
		k.CreateEpochUnbondingRecord(ctx, epochNumber)
	}

	if epochIdentifier == epochstypes.STRIDE_EPOCH {
		k.Logger(ctx).Info(utils.LogHeader("STRIDE EPOCH %d", epochNumber))

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
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// every epoch
	epochIdentifier := epochInfo.Identifier
	epochNumber := epochInfo.CurrentEpoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch end %s %d", epochIdentifier, epochNumber))
	if epochIdentifier == "day" {
		k.Logger(ctx).Info(fmt.Sprintf("Day %d Ending", epochNumber))
	}
}

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

// Set the withdrawal account address for each host zone
func (k Keeper) SetWithdrawalAddress(ctx sdk.Context) {
	k.Logger(ctx).Info("Setting Withdrawal Addresses...")

	setWithdrawalAddresses := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		err := k.SetWithdrawalAddressOnHost(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to set withdrawal address to %s on %s, err: %s", hostZone.WithdrawalAccount.Address, hostZone.ChainId, err))
		}
		return nil
	}

	k.IterateHostZones(ctx, setWithdrawalAddresses)
}

// Updates the redemption rate for each host zone
// The redemption rate equation is:
//   (Unbonded Balance + Staked Balance + Module Account Balance) / (stToken Supply)
func (k Keeper) UpdateRedemptionRates(ctx sdk.Context, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Updating Redemption Rates...")

	updateRedemptionRate := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		// Gather redemption rate components
		stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)).Amount.Int64()
		if stSupply == 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "No st%s in circulation - redemption rate is unchanged", hostZone.HostDenom))
			return nil
		}
		undelegatedBalance, err := k.GetUndelegatedBalance(hostZone, depositRecords)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get undelegated balance for host zone %s: %s", hostZone.ChainId, err.Error()))
			return err
		}
		stakedBalance, err := cast.ToInt64E(hostZone.StakedBal)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get staked balance for host zone %s: %s", hostZone.ChainId, err.Error()))
			return err
		}
		moduleAcctBalance, err := k.GetModuleAccountBalance(hostZone, depositRecords)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get module account balance for host zone %s: %s", hostZone.ChainId, err.Error()))
			return err
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"Redemption Rate Components - Undelegated Balance: %d, Staked Balance: %d, Module Account Balance: %d, stToken Supply: %d",
			undelegatedBalance, stakedBalance, moduleAcctBalance, stSupply))

		// Calculate the redemption rate
		redemptionRate := (sdk.NewDec(undelegatedBalance).Add(sdk.NewDec(stakedBalance)).Add(sdk.NewDec(moduleAcctBalance))).Quo(sdk.NewDec(stSupply))
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "New Redemption Rate: %d (vs Prev Rate: %d)", redemptionRate, hostZone.RedemptionRate))

		// Update the host zone
		hostZone.LastRedemptionRate = hostZone.RedemptionRate
		hostZone.RedemptionRate = redemptionRate
		k.SetHostZone(ctx, hostZone)

		return nil
	}

	// Iterate the zones and apply update the redemption rate for each
	k.IterateHostZones(ctx, updateRedemptionRate)
}

func (k Keeper) GetUndelegatedBalance(hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
	// filter to only the deposit records for the host zone with status DELEGATION_QUEUE
	UndelegatedDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return ((record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE || record.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS) && record.HostZoneId == hostZone.ChainId)
	})

	// sum the amounts of the deposit records
	var totalAmount int64
	for _, depositRecord := range UndelegatedDepositRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}

func (k Keeper) GetModuleAccountBalance(hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
	// filter to only the deposit records for the host zone with status DELEGATION
	ModuleAccountRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return (record.Status == recordstypes.DepositRecord_TRANSFER_QUEUE || record.Status == recordstypes.DepositRecord_TRANSFER_IN_PROGRESS) && record.HostZoneId == hostZone.ChainId
	})

	// sum the amounts of the deposit records
	totalAmount := int64(0)
	for _, depositRecord := range ModuleAccountRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}

func (k Keeper) ReinvestRewards(ctx sdk.Context) {
	k.Logger(ctx).Info("Reinvesting tokens...")

	for _, hostZone := range k.GetAllHostZone(ctx) {
		// only process host zones once withdrawal accounts are registered
		withdrawalIca := hostZone.WithdrawalAccount
		if withdrawalIca == nil {
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
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Updated withdrawal balance"))
	}
}
