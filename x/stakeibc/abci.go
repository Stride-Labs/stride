package stakeibc

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/Stride-Labs/stride/v6/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v6/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func HaltZone(ctx sdk.Context, k keeper.Keeper, hz types.HostZone, err error) {
	hz.Halted = true
	k.SetHostZone(ctx, hz)

	// set rate limit on stAsset
	stDenom := types.StAssetDenomFromHostZoneDenom(hz.HostDenom)
	k.RatelimitKeeper.AddDenomToBlacklist(ctx, stDenom)

	k.Logger(ctx).Error(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's RR is %s. ERR: %v", hz.GetChainId(), hz.RedemptionRate.String(), err.Error()))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHostZoneHalt,
			sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
			sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
		),
	)
}

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// invariant checks
	for _, hz := range k.GetAllHostZone(ctx) {
		// As a redundancy check, we store the module account balance for each host zone
		// as well as outstanding stToken supply. With these we calculate the implied
		// redemption rate in the EndBlocker and check that it is within the safety bounds.
		hz.BeginBlockModuleAcctHostDenomBal = bk.GetBalance(ctx, sdk.AccAddress(hz.Address), hz.IbcDenom).Amount
		hz.BeginBlockStDenomBal = bk.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hz.HostDenom)).Amount
		// CHECK: Does setting this while looping over all zones cause issues?
		k.SetHostZone(ctx, hz)

		rrSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hz)
		if !rrSafe {
			HaltZone(ctx, k, hz, err)
		}
	}
}

func EndBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// invariant checks
	for _, hz := range k.GetAllHostZone(ctx) {
		endBlockModuleAcctHostDenomBal := bk.GetBalance(ctx, sdk.AccAddress(hz.Address), hz.IbcDenom).Amount
		// CHECK: Supply only changes when Mint/Burn are called (not when IBC transfers occur)
		endBlockStDenomBal := bk.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hz.HostDenom)).Amount

		// Verify that the number of newly minted stTokens doesn't exceed the amount of tokens in the module account
		// (divided by the redemption rate)
		nativeTokensDepositedToModuleAcct := endBlockModuleAcctHostDenomBal.Sub(hz.BeginBlockModuleAcctHostDenomBal)
		stTokensMinted := endBlockStDenomBal.Sub(hz.BeginBlockStDenomBal)
		var impliedRedemptionRate sdk.Dec
		if stTokensMinted.GT(sdk.ZeroInt()) {
			impliedRedemptionRate = sdk.NewDecFromInt(nativeTokensDepositedToModuleAcct).Quo(sdk.NewDecFromInt(stTokensMinted))
		} else {
			continue
		}
		impliedRedemptionRateSafe, err := k.IsImpliedRedemptionRateWithinSafetyBounds(ctx, hz, impliedRedemptionRate)
		if !impliedRedemptionRateSafe {
			HaltZone(ctx, k, hz, err)
		}
	}
}
