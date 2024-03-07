package keeper

import (
	"encoding/json"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v19/utils"
	icaoracletypes "github.com/Stride-Labs/stride/v19/x/icaoracle/types"
	"github.com/Stride-Labs/stride/v19/x/staketia/types"
)

// Updates the redemption rate for each host zone
// At a high level, the redemption rate is equal to the amount of native tokens locked divided by the stTokens in existence.
// The equation is broken down further into the following sub-components:
//
//	Native Tokens Locked:
//	  1. Deposit Account Balance: tokens deposited from liquid stakes, that are still living on Stride
//	  2. Undelegated Balance:     tokens that are ready to be staked
//	                              (they're either currently in the delegation account or currently being transferred there)
//	  3. Delegated Balance:       Delegations on the host zone
//
//	StToken Amount:
//	  1. Total Supply of the stToken
//
//	Redemption Rate = (Deposit Account Balance + Undelegated Balance + Delegated Balance) / (stToken Supply)
//
// Note: Reinvested tokens are sent to the deposit account and are automatically included in this formula
func (k Keeper) UpdateRedemptionRate(ctx sdk.Context) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "Updating redemption rate"))

	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}

	// Get the number of stTokens from the supply
	stTokenSupply := k.bankKeeper.GetSupply(ctx, utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)).Amount
	if stTokenSupply.IsZero() {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"No st%s in circulation - redemption rate is unchanged", hostZone.NativeTokenDenom))
		return nil
	}

	// Get the balance of the deposit address
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid deposit address")
	}
	depositAccountBalance := k.bankKeeper.GetBalance(ctx, depositAddress, hostZone.NativeTokenIbcDenom)

	// Then add that to the sum of the delegation records to get the undelegated balance
	// Delegation records are only created once the tokens leave the deposit address
	// and the record is deleted once the tokens are delegated
	undelegatedBalance := sdkmath.ZeroInt()
	for _, delegationRecord := range k.GetAllActiveDelegationRecords(ctx) {
		undelegatedBalance = undelegatedBalance.Add(delegationRecord.NativeAmount)
	}

	// Finally, calculated the redemption rate as the native tokens locked divided by the stTokens
	nativeTokensLocked := depositAccountBalance.Amount.Add(undelegatedBalance).Add(hostZone.DelegatedBalance)
	if !nativeTokensLocked.IsPositive() {
		return errors.New("Non-zero stToken supply, yet the zero delegated and undelegated balance")
	}
	redemptionRate := sdk.NewDecFromInt(nativeTokensLocked).Quo(sdk.NewDecFromInt(stTokenSupply))

	// Set the old and update redemption rate on the host
	hostZone.LastRedemptionRate = hostZone.RedemptionRate
	hostZone.RedemptionRate = redemptionRate
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId, "Redemption rate updated from %v to %v",
		hostZone.LastRedemptionRate, hostZone.RedemptionRate))
	k.Logger(ctx).Info(utils.LogWithHostZone(types.CelestiaChainId,
		"Deposit Account Balance: %v, Undelegated Balance: %v, Delegated Balance: %v, StToken Supply: %v",
		depositAccountBalance.Amount, undelegatedBalance, hostZone.DelegatedBalance, stTokenSupply))

	return nil
}

// Checks whether the redemption rate has exceeded the inner or outer safety bounds
// and returns an error if so
func (k Keeper) CheckRedemptionRateExceedsBounds(ctx sdk.Context) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	redemptionRate := hostZone.RedemptionRate

	// Validate the safety bounds (e.g. that the inner is inside the outer)
	if err := hostZone.ValidateRedemptionRateBoundsInitalized(); err != nil {
		return err
	}

	// Check if the redemption rate is outside the outer bounds
	if redemptionRate.LT(hostZone.MinRedemptionRate) || redemptionRate.GT(hostZone.MaxRedemptionRate) {
		return types.ErrRedemptionRateOutsideSafetyBounds.Wrapf("redemption rate outside outer safety bounds")
	}

	// Check if it's outside the inner bounds
	if redemptionRate.LT(hostZone.MinInnerRedemptionRate) || redemptionRate.GT(hostZone.MaxInnerRedemptionRate) {
		return types.ErrRedemptionRateOutsideSafetyBounds.Wrapf("redemption rate outside inner safety bounds")
	}

	return nil
}

// Pushes a redemption rate update to the ICA oracle
func (k Keeper) PostRedemptionRateToOracles(ctx sdk.Context) error {
	if err := k.CheckRedemptionRateExceedsBounds(ctx); err != nil {
		return errorsmod.Wrapf(err, "preventing oracle update since redemption rate exceeded bounds")
	}

	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	redemptionRate := hostZone.RedemptionRate

	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
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
	k.icaOracleKeeper.QueueMetricUpdate(ctx, metricKey, metricValue, metricType, string(attributes))

	return nil
}
