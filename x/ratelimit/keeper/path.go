package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// Format the pathId as '{BaseDenom}/{ChannelId}
func FormatPathId(baseDenom string, channelId string) string {
	return fmt.Sprintf("%s/%s", baseDenom, channelId)
}

// Stores/Updates a path object in the store
func (k Keeper) SetPath(ctx sdk.Context, path types.Path) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PathKeyPrefix)

	pathKey := types.KeyPrefix(path.Id)
	pathValue := k.cdc.MustMarshal(&path)

	store.Set(pathKey, pathValue)
}

// Removes a path from the store using the Id
func (k Keeper) RemovePath(ctx sdk.Context, pathId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PathKeyPrefix)
	pathKey := types.KeyPrefix(pathId)
	store.Delete(pathKey)
}

// Grabs and returns a path from the store using the Id
func (k Keeper) GetPath(ctx sdk.Context, pathId string) (path types.Path, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PathKeyPrefix)

	pathKey := types.KeyPrefix(pathId)
	pathValue := store.Get(pathKey)

	if pathValue == nil || len(pathValue) == 0 {
		return path, false
	}

	k.cdc.MustUnmarshal(pathValue, &path)
	return path, true
}

// Returns all paths stored
func (k Keeper) GetAllPaths(ctx sdk.Context) []types.Path {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PathKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allPaths := []types.Path{}
	for ; iterator.Valid(); iterator.Next() {

		path := types.Path{}
		k.cdc.MustUnmarshal(iterator.Value(), &path)
		allPaths = append(allPaths, path)
	}

	return allPaths
}
