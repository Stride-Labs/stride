package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Stride-Labs/stride/v11/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

type StakeIbcPostDecorator struct {
	StakeIbcKeeper Keeper
}

func NewStakeIbcPostDecorator(stakeIbcDecorator Keeper) StakeIbcPostDecorator {
	return StakeIbcPostDecorator{
		StakeIbcKeeper: stakeIbcDecorator,
	}
}

// This posthandler will re-compute the redemption rate and check again the rate already saved in host zone
func (stakeIbcDec StakeIbcPostDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool,  next sdk.PostHandler) (sdk.Context, error) {
	if ctx.IsCheckTx() {
		return next(ctx, tx, simulate, success)
	}

	// Create a cache context to execute the posthandler such that
	// 1. If there is an error, then the cache context is discarded
	// 2. If there is no error, then the cache context is written to the main context with no gas consumed
	cacheCtx, write := ctx.CacheContext()
	// CacheCtx's by default _share_ their gas meter with the parent.
	// In our case, the cache ctx is given a new gas meter instance entirely,
	// so gas usage is not counted towards tx gas usage.
	//
	// 50M is chosen as a large enough number to ensure that the posthandler will not run out of gas,
	// but will eventually terminate in event of an accidental infinite loop with some gas usage.
	cacheCtx = cacheCtx.WithGasMeter(sdk.NewGasMeter(sdk.Gas(50_000_000)))

	moduleAddr := stakeIbcDec.StakeIbcKeeper.AccountKeeper.GetModuleAccount(ctx, stakeibctypes.ModuleName).GetAddress()
	hostzones := stakeIbcDec.StakeIbcKeeper.GetAllHostZone(cacheCtx)
	for _, hz := range hostzones {
		stSupplyAfter := stakeIbcDec.StakeIbcKeeper.bankKeeper.GetSupply(ctx, hz.HostDenom)
		stSupplyBefore, err := stakeIbcDec.StakeIbcKeeper.GetStSupply(ctx, hz)
		if err != nil {
			hz.Halted = true
			stakeIbcDec.StakeIbcKeeper.SetHostZone(ctx, hz)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeHostZoneHalt,
					sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
					sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
				),
			)
			continue
		}

		ibcDenomModuleAccountBalanceAfter := stakeIbcDec.StakeIbcKeeper.bankKeeper.GetBalance(ctx, moduleAddr, hz.IbcDenom)
		ibcDenomModuleAccountBalanceBefore, err := stakeIbcDec.StakeIbcKeeper.GetModuleAccountIbcBalance(ctx, hz)
		if err != nil {
			hz.Halted = true
			stakeIbcDec.StakeIbcKeeper.SetHostZone(ctx, hz)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeHostZoneHalt,
					sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
					sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
				),
			)
			continue
		}

		stSupplyDiff := stSupplyAfter.Amount.Sub(stSupplyBefore.Amount)
		ibcDenomBalanceDiff := ibcDenomModuleAccountBalanceAfter.Amount.Sub(ibcDenomModuleAccountBalanceBefore.Amount)

		expectedRedemptionRate := sdk.NewDecFromInt(ibcDenomBalanceDiff).Quo(sdk.NewDecFromInt(stSupplyDiff))
		if !expectedRedemptionRate.Equal(hz.RedemptionRate) {
			hz.Halted = true
			stakeIbcDec.StakeIbcKeeper.SetHostZone(ctx, hz)

			stakeIbcDec.StakeIbcKeeper.Logger(ctx).Error(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's recomputeRR is %s different from zone RR is %s.", hz.GetChainId(), expectedRedemptionRate.String(), hz.RedemptionRate.String()))
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeHostZoneHalt,
					sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
					sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
				),
			)
		}
	}
	write()

	return next(ctx, tx, simulate, success)
}

func(k Keeper) ComputeRedemptionRate(ctx sdk.Context, hz types.HostZone) sdk.Dec {
	stDenomSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(hz.HostDenom))
	fmt.Println("stDenomSupply", stDenomSupply)
	depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
	undelegatedBalance := k.GetUndelegatedBalance(hz, depositRecords)
	fmt.Println("undelegatedBalance", undelegatedBalance)
	stakedBalance := hz.StakedBal
	fmt.Println("stakedBalance", stakedBalance)
	moduleAcctBalance := k.GetModuleAccountBalance(hz, depositRecords)
	fmt.Println("moduleAcctBalance", moduleAcctBalance)
	redemptionRate := (sdk.NewDecFromInt(undelegatedBalance).Add(sdk.NewDecFromInt(stakedBalance)).Add(sdk.NewDecFromInt(moduleAcctBalance))).Quo(sdk.NewDecFromInt(stDenomSupply.Amount))
	return redemptionRate
}

func(k Keeper) CompareRedemptionRate(ctx sdk.Context, recomputeRedemptionRate sdk.Dec, hz types.HostZone) {
	fmt.Println("hz rate", hz.RedemptionRate)
	// If reComputeRedemptionRate not equal to current rate, hz should be halted
	if !recomputeRedemptionRate.Equal(hz.RedemptionRate) {
		hz.Halted = true
		k.SetHostZone(ctx, hz)

		// set rate limit on stAsset
		stDenom := types.StAssetDenomFromHostZoneDenom(hz.HostDenom)
		k.RatelimitKeeper.AddDenomToBlacklist(ctx, stDenom)

		k.Logger(ctx).Error(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's recomputeRR is %s different from zone RR is %s.", hz.GetChainId(), recomputeRedemptionRate.String(), hz.RedemptionRate.String()))
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeHostZoneHalt,
				sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
				sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
			),
		)
	}
}