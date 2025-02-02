package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/auction/types"
)

type Keeper struct {
	cdc             codec.Codec
	storeKey        storetypes.StoreKey
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	icqoracleKeeper types.IcqOracleKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	icqoracleKeeper types.IcqOracleKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		icqoracleKeeper: icqoracleKeeper,
	}
}

// SetAuction stores auction info for a token
func (k Keeper) SetAuction(ctx sdk.Context, auction *types.Auction) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionPrefix)
	key := []byte(auction.Name)

	bz, err := k.cdc.Marshal(auction)
	if err != nil {
		return fmt.Errorf("error setting auction for name='%s': %w", auction.Name, err)
	}

	store.Set(key, bz)
	return nil
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

	// Get the appropriate auctionBidHandler for the auction type
	auctionBidHandler, exists := bidHandlers[auction.Type]
	if !exists {
		return fmt.Errorf("unsupported auction type: %s", auction.Type)
	}

	// Call the handler
	return auctionBidHandler(ctx, k, auction, bid)
}

// GetStoreKey returns the store key
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}
