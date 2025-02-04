package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	params, err := k.GetParams(ctx)
	if err != nil {
		ctx.Logger().Error("failed to get icqoracle params: %w", err)
		return
	}

	currentTime := ctx.BlockTime()

	for _, tokenPrice := range k.GetAllTokenPrices(ctx) {
		// Get last update time for this token
		lastUpdate := tokenPrice.LastRequestTime
		isNewToken := lastUpdate.IsZero()
		updateIntervalPassed := currentTime.Sub(lastUpdate) >= time.Second*time.Duration(params.UpdateIntervalSec)

		// If never updated or update interval has passed, submit a new query for the price
		// If a query was already in progress, it will be replaced with this new one that will
		// will have the same query ID
		if isNewToken || updateIntervalPassed {
			if err := k.SubmitOsmosisClPoolICQ(ctx, tokenPrice); err != nil {
				ctx.Logger().Error(
					"failed to submit Osmosis CL pool ICQ baseToken='%s' quoteToken='%s' poolId=%d: %w",
					tokenPrice.BaseDenom,
					tokenPrice.QuoteDenom,
					tokenPrice.OsmosisPoolId,
					err,
				)
				continue
			}
		}
	}
}
