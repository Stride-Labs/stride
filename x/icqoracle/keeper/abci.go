package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	// Get all token prices
	params := k.GetParams(ctx)

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
				// TODO handle error, maybe log it
				continue
			}
		}
	}
}
