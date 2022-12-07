package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func FormatPathId(denom string, channelId string) string {
	return fmt.Sprintf("%s-%s", denom, channelId)
}

func (k Keeper) AddPath(ctx sdk.Context, path types.Path) string {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefix(types.PathKey))

	pathId := FormatPathId(path.Denom, path.ChannelId)
	pathKey := types.KeyPrefix(pathId)
	path.Id = pathId

	pathBytes := k.cdc.MustMarshal(&path)

	prefixStore.Set(pathKey, pathBytes)

	return pathId
}

func (k Keeper) RemovePath(ctx sdk.Context, pathId string) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefix(types.PathKey))

	pathKey := types.KeyPrefix(pathId)

	prefixStore.Delete(pathKey)
}

func (k Keeper) GetPath(ctx sdk.Context, pathId string) (path types.Path, found bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefix(types.PathKey))

	pathKey := types.KeyPrefix(pathId)
	pathBytes := prefixStore.Get(pathKey)

	path = types.Path{}
	err := proto.Unmarshal(pathBytes, &path)
	if err != nil {
		return types.Path{}, false
	}

	return path, true
}

func (k Keeper) GetAllPaths(ctx sdk.Context) []types.Path {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefix(types.PathKey))

	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	allPaths := []types.Path{}
	for ; iterator.Valid(); iterator.Next() {

		path := types.Path{}
		err := proto.Unmarshal(iterator.Value(), &path)
		if err != nil {
			panic(err)
		}

		allPaths = append(allPaths, path)
	}

	return allPaths
}
