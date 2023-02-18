package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
)

// Before each hour epoch, check if any of the rate limits have expired,
//  and reset them if they have
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	if epochInfo.Identifier == epochstypes.HOUR_EPOCH {
		epochHour := uint64(epochInfo.CurrentEpoch)

		for _, rateLimit := range k.GetAllRateLimits(ctx) {
			if epochHour%rateLimit.Quota.DurationHours == 0 {
				err := k.ResetRateLimit(ctx, rateLimit.Path.Denom, rateLimit.Path.ChannelId)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Unable to reset quota for Denom: %s, ChannelId: %s", rateLimit.Path.Denom, rateLimit.Path.ChannelId))
				}
			}
		}
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {}

type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}
