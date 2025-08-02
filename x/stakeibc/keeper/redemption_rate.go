package keeper

import (
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/utils"
	icaoracletypes "github.com/Stride-Labs/stride/v28/x/icaoracle/types"
	recordstypes "github.com/Stride-Labs/stride/v28/x/records/types"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

// Updates the redemption rate for each host zone
// At a high level, the redemption rate is equal to the amount of native tokens locked divided by the stTokens in existence.
// The equation is broken down further into the following sub-components:
//
//	   Native Tokens Locked:
//	     1. Deposit Account Balance: native tokens deposited from liquid stakes, that are still living on Stride
//	     2. Undelegated Balance:     native tokens that have been transferred to the host zone, but have not been delegated yet
//	     3. Tokenized Delegations:   Delegations inherent in LSM Tokens that have not yet been converted to native stake
//	     4. Native Delegations:      Delegations either from native tokens, or LSM Tokens that have been detokenized
//	  StToken Amount:
//	     1. Total Supply of the stToken
//
//	Redemption Rate =
//	(Deposit Account Balance + Undelegated Balance + Tokenized Delegation + Native Delegation) / (stToken Supply)
func (k Keeper) UpdateRedemptionRates(ctx sdk.Context, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Updating Redemption Rates...")

	// Update the redemption rate for each host zone
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		k.UpdateRedemptionRateForHostZone(ctx, hostZone, depositRecords)
	}
}

func (k Keeper) UpdateRedemptionRateForHostZone(ctx sdk.Context, hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) {
	// Gather redemption rate components
	stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)).Amount
	if stSupply.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"No st%s in circulation - redemption rate is unchanged", hostZone.HostDenom))
		return
	}

	depositAccountBalance := k.GetDepositAccountBalance(hostZone.ChainId, depositRecords)
	undelegatedBalance := k.GetUndelegatedBalance(hostZone.ChainId, depositRecords)
	tokenizedDelegation := k.GetTotalTokenizedDelegations(ctx, hostZone)
	nativeDelegation := sdkmath.LegacyNewDecFromInt(hostZone.TotalDelegations)

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Redemption Rate Components - Deposit Account Balance: %v, Undelegated Balance: %v, "+
			"LSM Delegated Balance: %v, Native Delegations: %v, stToken Supply: %v",
		depositAccountBalance, undelegatedBalance, tokenizedDelegation,
		nativeDelegation, stSupply))

	// Calculate the redemption rate
	nativeTokensLocked := depositAccountBalance.Add(undelegatedBalance).Add(tokenizedDelegation).Add(nativeDelegation)
	redemptionRate := nativeTokensLocked.Quo(sdkmath.LegacyNewDecFromInt(stSupply))

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"New Redemption Rate: %v (vs Prev Rate: %v)", redemptionRate, hostZone.RedemptionRate))

	// Update the host zone
	hostZone.LastRedemptionRate = hostZone.RedemptionRate
	hostZone.RedemptionRate = redemptionRate
	k.SetHostZone(ctx, hostZone)

	// If the redemption rate is outside of safety bounds, exit so the redemption rate is not pushed to the oracle
	redemptionRateSafe, _ := k.IsRedemptionRateWithinSafetyBounds(ctx, hostZone)
	if !redemptionRateSafe {
		return
	}

	// Otherwise, submit the redemption rate to the oracle
	if err := k.PostRedemptionRateToOracles(ctx, hostZone.HostDenom, redemptionRate); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Unable to send redemption rate to oracle: %s", err.Error()))
		return
	}
}

// Determine the deposit account balance, representing native tokens that have been deposited
// from liquid stakes, but have not yet been transferred to the host
func (k Keeper) GetDepositAccountBalance(chainId string, depositRecords []recordstypes.DepositRecord) sdkmath.LegacyDec {
	// sum on deposit records with status TRANSFER_QUEUE or TRANSFER_IN_PROGRESS
	totalAmount := sdkmath.ZeroInt()
	for _, depositRecord := range depositRecords {
		transferStatus := (depositRecord.Status == recordstypes.DepositRecord_TRANSFER_QUEUE ||
			depositRecord.Status == recordstypes.DepositRecord_TRANSFER_IN_PROGRESS)

		if depositRecord.HostZoneId == chainId && transferStatus {
			totalAmount = totalAmount.Add(depositRecord.Amount)
		}
	}

	return sdkmath.LegacyNewDecFromInt(totalAmount)
}

// Determine the undelegated balance from the deposit records queued for staking
func (k Keeper) GetUndelegatedBalance(chainId string, depositRecords []recordstypes.DepositRecord) sdkmath.LegacyDec {
	// sum on deposit records with status DELEGATION_QUEUE or DELEGATION_IN_PROGRESS
	totalAmount := sdkmath.ZeroInt()
	for _, depositRecord := range depositRecords {
		delegationStatus := (depositRecord.Status == recordstypes.DepositRecord_DELEGATION_QUEUE ||
			depositRecord.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS)

		if depositRecord.HostZoneId == chainId && delegationStatus {
			totalAmount = totalAmount.Add(depositRecord.Amount)
		}
	}

	return sdkmath.LegacyNewDecFromInt(totalAmount)
}

// Returns the total delegated balance that's stored in LSM tokens
// This is used for the redemption rate calculation
//
// The relevant tokens are identified by the deposit records in status "DEPOSIT_PENDING"
// "DEPOSIT_PENDING" means the liquid staker's tokens have not been sent to Stride yet
// so they should *not* be included in the redemption rate. All other statuses indicate
// the LSM tokens have been deposited and should be included in the final calculation
//
// Each LSM token represents a delegator share so the validator's shares to tokens rate
// must be used to denominate it's value in native tokens
func (k Keeper) GetTotalTokenizedDelegations(ctx sdk.Context, hostZone types.HostZone) sdkmath.LegacyDec {
	total := sdkmath.ZeroInt()
	for _, deposit := range k.RecordsKeeper.GetLSMDepositsForHostZone(ctx, hostZone.ChainId) {
		if deposit.Status != recordstypes.LSMTokenDeposit_DEPOSIT_PENDING {
			validator, _, found := GetValidatorFromAddress(hostZone.Validators, deposit.ValidatorAddress)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("Validator %s found in LSMTokenDeposit but no longer exists", deposit.ValidatorAddress))
				continue
			}
			liquidStakedShares := deposit.Amount
			liquidStakedTokens := sdkmath.LegacyNewDecFromInt(liquidStakedShares).Mul(validator.SharesToTokensRate)
			total = total.Add(liquidStakedTokens.TruncateInt())
		}
	}

	return sdkmath.LegacyNewDecFromInt(total)
}

// Pushes a redemption rate update to the ICA oracle
func (k Keeper) PostRedemptionRateToOracles(ctx sdk.Context, hostDenom string, redemptionRate sdkmath.LegacyDec) error {
	stDenom := types.StAssetDenomFromHostZoneDenom(hostDenom)
	attributes, err := json.Marshal(icaoracletypes.RedemptionRateAttributes{
		SttokenDenom: stDenom,
	})
	if err != nil {
		return err
	}

	// Metric Key is of format: {stToken}_redemption_rate
	metricKey := fmt.Sprintf("%s_%s", stDenom, icaoracletypes.MetricType_RedemptionRate)
	metricValue := redemptionRate.String()
	metricType := icaoracletypes.MetricType_RedemptionRate
	k.ICAOracleKeeper.QueueMetricUpdate(ctx, metricKey, metricValue, metricType, string(attributes))

	return nil
}
