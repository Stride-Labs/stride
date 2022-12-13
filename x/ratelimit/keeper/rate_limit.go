package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// QUESTION FOR REVIEWER: Is this the right way to do a composite key?
func GetRateLimitItemKey(denom string, channelId string) []byte {
	return append(types.KeyPrefix(denom), types.KeyPrefix(channelId)...)
}

func CheckRateLimit(direction types.PacketDirection, packet exported.PacketI) error {
	// TODO
	return nil
}

// Stores/Updates a rate limit object in the store
func (k Keeper) SetRateLimit(ctx sdk.Context, rateLimit types.RateLimit) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	rateLimitKey := GetRateLimitItemKey(rateLimit.Path.Denom, rateLimit.Path.ChannelId)
	rateLimitValue := k.cdc.MustMarshal(&rateLimit)

	store.Set(rateLimitKey, rateLimitValue)
}

// Removes a rate limit object from the store using the PathId
func (k Keeper) RemoveRateLimit(ctx sdk.Context, denom string, channelId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)
	rateLimitKey := GetRateLimitItemKey(denom, channelId)
	store.Delete(rateLimitKey)
}

// Grabs and returns a rate limit object from the store using the PathId
func (k Keeper) GetRateLimit(ctx sdk.Context, denom string, channelId string) (rateLimit types.RateLimit, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	rateLimitKey := GetRateLimitItemKey(denom, channelId)
	rateLimitValue := store.Get(rateLimitKey)

	if rateLimitValue == nil || len(rateLimitValue) == 0 {
		return rateLimit, false
	}

	k.cdc.MustUnmarshal(rateLimitValue, &rateLimit)
	return rateLimit, true
}

// Returns all rate limits stored
func (k Keeper) GetAllRateLimits(ctx sdk.Context) []types.RateLimit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RateLimitKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allRateLimits := []types.RateLimit{}
	for ; iterator.Valid(); iterator.Next() {

		rateLimit := types.RateLimit{}
		k.cdc.MustUnmarshal(iterator.Value(), &rateLimit)
		allRateLimits = append(allRateLimits, rateLimit)
	}

	return allRateLimits
}
