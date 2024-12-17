package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/strdburner/types"
)

func EndBlocker(ctx sdk.Context, k Keeper) {
	strdBurnerAddress := k.GetStrdBurnerAddress()

	strdBalance := k.bankKeeper.GetBalance(ctx, strdBurnerAddress, "ustrd")

	if strdBalance.IsPositive() {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(strdBalance))
		if err != nil {
			k.Logger(ctx).Error("unable to burn %s: %w", strdBalance.String(), err)
		} else {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeBurn,
					sdk.NewAttribute(types.AttributeAmount, strdBalance.String()),
				),
			)
		}
	}
}
