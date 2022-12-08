package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// Format the pathId as '{BaseDenom}_{ChannelId}
func FormatPathId(baseDenom string, channelId string) string {
	return fmt.Sprintf("%s_%s", baseDenom, channelId)
}

// Add a new path to the store
// The Id is generated from the TraceDenom (e.g. ibc/...) and ChannelId
func (k Keeper) AddPath(ctx sdk.Context, traceDenom string, channelId string) (pathId string, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PathKey))

	// If the trace denom DOES NOT have a ibc/ prefix, use that denom for the base denom
	// If the trace denom DOES have an ibc/ prefix, determine it's base denom from the hash
	var baseDenom string
	if !strings.HasPrefix(traceDenom, "ibc/") {
		baseDenom = traceDenom
	} else {
		// Convert the hash to bytes and use it to look up the denom trace
		traceHex, err := transfertypes.ParseHexHash(traceDenom)
		if err != nil {
			return "", err
		}
		denomTrace, found := k.transferKeeper.GetDenomTrace(ctx, traceHex)
		if !found {
			return "", fmt.Errorf("Unable to determine denom trace from hash %s", traceDenom)
		}
		baseDenom = denomTrace.BaseDenom
	}

	// pathId is of the form '{BaseDenom}_{ChannelId}'
	pathId = FormatPathId(baseDenom, channelId)
	pathKey := types.KeyPrefix(pathId)

	pathValue := k.cdc.MustMarshal(&types.Path{
		Id:         pathId,
		BaseDenom:  baseDenom,
		TraceDenom: traceDenom,
		ChannelId:  channelId,
	})

	store.Set(pathKey, pathValue)

	return pathId, nil
}

// Stores/Updates a path object in the store
func (k Keeper) SetPath(ctx sdk.Context, path types.Path) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PathKey))

	pathKey := types.KeyPrefix(path.Id)
	pathValue := k.cdc.MustMarshal(&path)

	store.Set(pathKey, pathValue)
}

// Removes a path from the store using the Id
func (k Keeper) RemovePath(ctx sdk.Context, pathId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PathKey))
	pathKey := types.KeyPrefix(pathId)
	store.Delete(pathKey)
}

// Grabs and returns a path from the store using the Id
func (k Keeper) GetPath(ctx sdk.Context, pathId string) (path types.Path, found bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefix(types.PathKey))

	pathKey := types.KeyPrefix(pathId)
	pathValue := prefixStore.Get(pathKey)

	path = types.Path{}
	err := proto.Unmarshal(pathValue, &path)
	if err != nil {
		return types.Path{}, false
	}

	return path, true
}

// Returns all paths stored
func (k Keeper) GetAllPaths(ctx sdk.Context) []types.Path {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PathKey))

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
