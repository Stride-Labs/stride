package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// Before each hour epoch, check if any of the rate limits have expired,
//  and reset them if they have
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	if epochInfo.Identifier == epochstypes.HOUR_EPOCH {
		epochHour := uint64(epochInfo.CurrentEpoch)

		for _, rateLimit := range k.GetAllRateLimits(ctx) {
			if rateLimit.Quota.DurationHours%epochHour == 0 {
				k.ResetRateLimit(ctx, rateLimit)
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
