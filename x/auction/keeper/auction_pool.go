package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// StartNewAuctionPool can be called by external modules like stakeibc
// In the future instead of a single set of pools, should be prefixed by hostzone
func (k Keeper) StartNewAuctionPool(ctx sdk.Context, properties types.AuctionPoolProperties) {
	pool := types.AuctionPool{}

	pool.PoolProperties = &properties
	pool.LatestAuction = &types.Auction{}

	k.AppendAuctionPool(ctx, pool)
}

// StartNewAuction updates the relevant auctionPool in the store to have start and end blocks running now
func (k Keeper) StartNewAuction(ctx sdk.Context, auctionPool types.AuctionPool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AuctionPoolKey))

	algorithm := types.AuctionType_SEALEDBID // default for now
	allowedAlgorithms := auctionPool.GetPoolProperties().GetAllowedAlgorithms()
	if len(allowedAlgorithms) == 1 {
		algorithm = allowedAlgorithms[0]
	}
	auctionPool.LatestAuction = &types.Auction{}
	auctionPool.GetLatestAuction().Algorithm = algorithm

	addr, _ := sdk.AccAddressFromBech32(auctionPool.GetPoolProperties().GetPoolAddress())
	coins := k.bankKeeper.SpendableCoins(ctx, addr)
	//very temporary, normally the Denom properties are fixed and used to look for which coins match
	auctionPool.GetPoolProperties().SupplyDenom = coins[0].Denom
	auctionPool.GetPoolProperties().BidDenom = coins[0].Denom

	switch algorithm {
	case types.AuctionType_ASCENDING:
		auction := types.AscendingAuction{}
		auctionPool.GetLatestAuction().XAscendingAuction = &types.Auction_AscendingAuction{&auction}
		properties := auctionPool.GetPoolProperties().GetDefaultAscendingAuctionProperties()
		properties.Supply = coins[0].Amount.Uint64()
		ctx.Logger().Info(fmt.Sprintf("[auction] Coins in the auction address %d %s", properties.GetSupply(), auctionPool.GetPoolProperties().GetSupplyDenom()))
		auction.CreateNew(ctx, properties)
		auction.PoolProperties = auctionPool.GetPoolProperties()
	case types.AuctionType_DESCENDING:
		auction := types.DescendingAuction{}
		auctionPool.GetLatestAuction().XDescendingAuction = &types.Auction_DescendingAuction{&auction}
		properties := auctionPool.GetPoolProperties().GetDefaultDescendingAuctionProperties()
		properties.Supply = coins[0].Amount.Uint64()
		ctx.Logger().Info(fmt.Sprintf("[auction] Coins in the auction address %d %s", properties.GetSupply(), auctionPool.GetPoolProperties().GetSupplyDenom()))
		auction.CreateNew(ctx, properties)
		auction.PoolProperties = auctionPool.GetPoolProperties()
	case types.AuctionType_SEALEDBID:
		auction := types.SealedBidAuction{}
		auctionPool.GetLatestAuction().XSealedBidAuction = &types.Auction_SealedBidAuction{&auction}
		properties := auctionPool.GetPoolProperties().GetDefaultSealedBidAuctionProperties()
		properties.Supply = coins[0].Amount.Uint64()
		ctx.Logger().Info(fmt.Sprintf("[auction] Coins in the auction address %d %s", properties.GetSupply(), auctionPool.GetPoolProperties().GetSupplyDenom()))
		auction.CreateNew(ctx, properties)
		auction.PoolProperties = auctionPool.GetPoolProperties()
	}

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
