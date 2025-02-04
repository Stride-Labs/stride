package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	if err := k.RefreshTokenPrices(ctx); err != nil {
		ctx.Logger().Error(fmt.Sprintf("failed to refresh token prices: %s", err.Error()))
	}
}
