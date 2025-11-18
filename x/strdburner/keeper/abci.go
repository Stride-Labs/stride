package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v30/x/strdburner/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) {
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

	// Update TotalStrdBurned and ProtocolStrdBurned
	k.IncrementProtocolStrdBurned(ctx, strdBalance.Amount)

	// Emit burn event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.AttributeAmount, strdBalance.String()),
		),
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProtocolBurn,
			sdk.NewAttribute(types.AttributeAmount, strdBalance.String()),
		),
	)
}
