package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	params, err := k.GetParams(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to fetch params")
		return
	}

	for _, tokenPrice := range k.GetAllTokenPrices(ctx) {
		if err := k.RefreshTokenPrice(ctx, tokenPrice, params.UpdateIntervalSec); err != nil {
			ctx.Logger().Error(fmt.Sprintf("failed to refresh token price: %s", err.Error()))
			continue
		}
	}
}
