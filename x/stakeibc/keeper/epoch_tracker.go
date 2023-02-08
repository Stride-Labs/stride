package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// SetEpochTracker set a specific epochTracker in the store from its index
func (k Keeper) SetEpochTracker(ctx sdk.Context, epochTracker types.EpochTracker) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	b := k.cdc.MustMarshal(&epochTracker)
	store.Set(types.EpochTrackerKey(
		epochTracker.EpochIdentifier,
	), b)
}

// GetEpochTracker returns a epochTracker from its index
func (k Keeper) GetEpochTracker(
	ctx sdk.Context,
	epochIdentifier string,
) (val types.EpochTracker, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))

	b := store.Get(types.EpochTrackerKey(
		epochIdentifier,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEpochTracker removes a epochTracker from the store
func (k Keeper) RemoveEpochTracker(
	ctx sdk.Context,
	epochIdentifier string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	store.Delete(types.EpochTrackerKey(
		epochIdentifier,
	))
}

// GetAllEpochTracker returns all epochTracker
func (k Keeper) GetAllEpochTracker(ctx sdk.Context) (list []types.EpochTracker) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochTrackerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochTracker
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
