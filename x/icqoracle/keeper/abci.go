package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/utils"
	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	params := k.GetParams(ctx)

	for _, tokenPrice := range k.GetAllTokenPrices(ctx) {
		if err := k.RefreshTokenPrice(ctx, tokenPrice, params.UpdateIntervalSec); err != nil {
			ctx.Logger().Error(fmt.Sprintf("failed to refresh token price: %s", err.Error()))
			continue
		}
	}
}

// Refreshes the price of a token (if applicable)
func (k Keeper) RefreshTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice, updateIntervalSec uint64) error {
	// Get last update time for this token
	currentTime := ctx.BlockTime()
	lastUpdate := tokenPrice.LastRequestTime
	isNewToken := lastUpdate.IsZero()
	updateIntervalPassed := currentTime.Sub(lastUpdate) >= time.Second*time.Duration(utils.UintToInt(updateIntervalSec))

	// If the update interval has not passed, don't update
	if !isNewToken && !updateIntervalPassed {
		return nil
	}

	// If never updated or update interval has passed, submit a new query for the price
	// If a query was already in progress, it will be replaced with a new one that will
	// have the same query ID
	if err := k.SubmitOsmosisPriceICQ(ctx, tokenPrice); err != nil {
		return errorsmod.Wrapf(err,
			"failed to submit Osmosis CL pool ICQ baseToken='%s' quoteToken='%s' poolId='%d'",
			tokenPrice.BaseDenom,
			tokenPrice.QuoteDenom,
			tokenPrice.OsmosisPoolId)
	}

	return nil
}
