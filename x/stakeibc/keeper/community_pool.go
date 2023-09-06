package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ibc transfers tokens from the foreign hub community pool deposit ICA address onto Stride hub
func (k Keeper) IBCTransferCommunityPoolICATokensToStride(ctx sdk.Context, token sdk.Coin) error {
	k.Logger(ctx).Info("Transfering tokens from community pool to Stride...")

	return nil
}
