package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/strdburner/types"
)

func EndBlocker(ctx sdk.Context, k Keeper) {
	strdBurnerAddress := k.GetStrdBurnerAddress()

	// Get STRD balance
	strdBalance := k.bankKeeper.GetBalance(ctx, strdBurnerAddress, "ustrd")

	// Exit early if nothing to burn
	if strdBalance.IsZero() {
		return
	}

	// Burn all STRD balance
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(strdBalance))
	if err != nil {
		k.Logger(ctx).Error("unable to burn %s: %w", strdBalance.String(), err)
		return
	}

	// Update TotalStrdBurned
	currentTotalBurned := k.GetTotalStrdBurned(ctx)
	newTotalBurned := currentTotalBurned.Add(strdBalance.Amount)
	k.SetTotalStrdBurned(ctx, newTotalBurned)

	// Emit burn event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.AttributeAmount, strdBalance.String()),
		),
	)
}
