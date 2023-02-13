package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// StartNewAuction updates the relevant auctionPool in the store to have start and end blocks running now
func (k Keeper) StartNewAuction(ctx sdk.Context, auctionPool types.AuctionPool, auctionSettings interface{}) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))

	now := cast.ToUint64(ctx.BlockHeight())

	switch auctionSettings.(type) {
	case types.SealedBidAuction:
		auction := auctionPool.GetLatestAuction().GetAuction().(*types.SealedBidAuction)
		auction.FirstBlock = now
		auction.LastBlock = auction.FirstBlock + auction.GetAuctionDuration()
		auction.RevealBlock = auction.LastBlock + auction.GetRevealDuration()
	default:

	}
	// TODO: check the amount of coin in the address of this pool and update that
	// TODO: also take in an auction type and if it is sealed, then also update the revealBlock

	updated := k.cdc.MustMarshal(&auctionPool)
	store.Set(GetAuctionPoolIDBytes(auctionPool.Id), updated)
}

// GetAuctionPoolCount get the total number of auctionPool
func (k Keeper) GetAuctionPoolCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.AuctionPoolCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetAuctionPoolCount set the total number of auctionPool
func (k Keeper) SetAuctionPoolCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.AuctionPoolCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendAuctionPool appends a auctionPool in the store with a new id and update the count
func (k Keeper) AppendAuctionPool(
	ctx sdk.Context,
	auctionPool types.AuctionPool,
) uint64 {
	// Create the auctionPool
	count := k.GetAuctionPoolCount(ctx)

	// Set the ID of the appended value
	auctionPool.Id = count

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))
	appendedValue := k.cdc.MustMarshal(&auctionPool)
	store.Set(GetAuctionPoolIDBytes(auctionPool.Id), appendedValue)

	// Update auctionPool count
	k.SetAuctionPoolCount(ctx, count+1)

	return count
}

// SetAuctionPool set a specific auctionPool in the store
func (k Keeper) SetAuctionPool(ctx sdk.Context, auctionPool types.AuctionPool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))
	b := k.cdc.MustMarshal(&auctionPool)
	store.Set(GetAuctionPoolIDBytes(auctionPool.Id), b)
}

// GetAuctionPool returns a auctionPool from its id
func (k Keeper) GetAuctionPool(ctx sdk.Context, id uint64) (val types.AuctionPool, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))
	b := store.Get(GetAuctionPoolIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveAuctionPool removes a auctionPool from the store
func (k Keeper) RemoveAuctionPool(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))
	store.Delete(GetAuctionPoolIDBytes(id))
}

// GetAllAuctionPool returns all auctionPool
func (k Keeper) GetAllAuctionPool(ctx sdk.Context) (list []types.AuctionPool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.AuctionPool
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetAuctionPoolIDBytes returns the byte representation of the ID
func GetAuctionPoolIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetAuctionPoolIDFromBytes returns ID in uint64 format from a byte array
func GetAuctionPoolIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
