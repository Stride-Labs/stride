package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	// Get all token prices
	params, err := k.GetParams(ctx)
	if err != nil {
		// Can't really do anything but log
		// A panic would halt the chain
		ctx.Logger().Error("failed to get icqoracle params: %w", err)
		return
	}

	currentTime := ctx.BlockTime()

	// Iterate through each token price
	tokenPrices := k.GetAllTokenPrices(ctx)
	for _, tokenPrice := range tokenPrices {
		// Get last update time for this token
		lastUpdate := tokenPrice.UpdatedAt

		// If never updated or update interval has passed
		if lastUpdate.IsZero() || !tokenPrice.QueryInProgress && currentTime.Sub(lastUpdate) >= time.Second*time.Duration(params.UpdateIntervalSec) {
			// Update price for this specific token
			err := k.SubmitOsmosisClPoolICQ(ctx, tokenPrice)
			if err != nil {
				// Can't really do anything but log
				ctx.Logger().Error(
					"failed to submit Osmosis CL pool ICQ error='%w' baseToken='%s' quoteToken='%s' poolId='%s'",
					err,
					tokenPrice.BaseDenom,
					tokenPrice.QuoteDenom,
					tokenPrice.OsmosisPoolId,
				)
				continue
			}
		}
	}
}
