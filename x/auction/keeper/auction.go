package keeper

import (
	"fmt"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/auction/types"
)

// SetAuction stores auction info for a token
func (k Keeper) SetAuction(ctx sdk.Context, auction *types.Auction) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionPrefix)
	key := []byte(auction.Name)
	bz := k.cdc.MustMarshal(auction)
	store.Set(key, bz)
}

// GetAuction retrieves auction info for a token
func (k Keeper) GetAuction(ctx sdk.Context, name string) (*types.Auction, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionPrefix)
	key := []byte(name)

	bz := store.Get(key)
	if bz == nil {
		return &types.Auction{}, fmt.Errorf("auction not found for denom '%s'", name)
	}

	var auction types.Auction
	if err := k.cdc.Unmarshal(bz, &auction); err != nil {
		return &types.Auction{}, fmt.Errorf("error retrieving auction for denom '%s': %w", auction.SellingDenom, err)
	}

	return &auction, nil
}

// GetAllAuctions retrieves all stored auctions
func (k Keeper) GetAllAuctions(ctx sdk.Context) []types.Auction {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	auctions := []types.Auction{}
	for ; iterator.Valid(); iterator.Next() {
		var auction types.Auction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)
		auctions = append(auctions, auction)
	}

	return auctions
}

// PlaceBid places an auction bid and executes it based on the auction type
func (k Keeper) PlaceBid(ctx sdk.Context, bid *types.MsgPlaceBid) error {
	// Get auction
	auction, err := k.GetAuction(ctx, bid.AuctionName)
	if err != nil {
		return fmt.Errorf("cannot get auction for name='%s': %w", bid.AuctionName, err)
	}

	if bid.PaymentTokenAmount.LT(auction.MinBidAmount) {
		return fmt.Errorf("payment bid amount '%s' is less than the minimum bid '%s' amount for auction '%s'", bid.PaymentTokenAmount.String(), auction.MinBidAmount.String(), bid.AuctionName)
	}

	// Get the appropriate auctionBidHandler for the auction type
	auctionBidHandler, exists := bidHandlers[auction.Type]
	if !exists {
		return fmt.Errorf("unsupported auction type: %s", auction.Type)
	}

	// Call the handler
	return auctionBidHandler(ctx, k, auction, bid)
}
