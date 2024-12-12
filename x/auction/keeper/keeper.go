package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetAuction stores auction info for a token
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) error {
	store := ctx.KVStore(k.storeKey)
	key := types.AuctionKey(auction.Denom)
	bz, err := k.cdc.Marshal(&auction)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// GetAuction retrieves auction info for a token
func (k Keeper) GetAuction(ctx sdk.Context, denom string) (types.Auction, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.AuctionKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return types.Auction{}, fmt.Errorf("auction not found for denom='%s'", denom)
	}

	var auction types.Auction
	if err := k.cdc.Unmarshal(bz, &auction); err != nil {
		return types.Auction{}, err
	}

	return auction, nil
}

// GetAllAuctions retrieves all stored auctions
func (k Keeper) GetAllAuctions(ctx sdk.Context) []types.Auction {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.KeyAuctionPrefix))
	defer iterator.Close()

	auctions := []types.Auction{}
	for ; iterator.Valid(); iterator.Next() {
		var auction types.Auction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)
		auctions = append(auctions, auction)
	}

	return auctions
}

// GetStats retrieves stats info for an auction
func (k Keeper) GetStats(ctx sdk.Context, denom string) (types.Stats, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.StatsKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return types.Stats{}, fmt.Errorf("Stats not found for auction denom='%s'", denom)
	}

	var stats types.Stats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		return types.Stats{}, err
	}

	return stats, nil
}

// GetAllStats retrieves all stored auctions stats
func (k Keeper) GetAllStats(ctx sdk.Context) []types.Stats {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.KeyStatsPrefix))
	defer iterator.Close()

	allStats := []types.Stats{}
	for ; iterator.Valid(); iterator.Next() {
		var auctionStats types.Stats
		k.cdc.MustUnmarshal(iterator.Value(), &auctionStats)
		allStats = append(allStats, auctionStats)
	}

	return allStats
}

// SetAuction stores auction info for a token
func (k Keeper) PlaceBid(ctx sdk.Context, bid *types.MsgPlaceBid) error {
	// Get auction
	auction, err := k.GetAuction(ctx, bid.TokenDenom)
	if err != nil {
		return fmt.Errorf("cannot get auction for denom='%s': %w", bid.TokenDenom, err)
	}

	// Get token amount being auctioned off
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	balance := k.bankKeeper.GetBalance(ctx, moduleAddr, auction.Denom)
	tokenAmount := balance.Amount

	// Verify auction has enough tokens to service the bid amount
	if bid.UtokenAmount.GT(tokenAmount) {
		return fmt.Errorf("bid wants %s%s but auction has only %s%s",
			bid.UtokenAmount.String(),
			bid.TokenDenom,
			tokenAmount.String(),
			bid.TokenDenom,
		)
	}

	// TODO continue

	return nil
}
