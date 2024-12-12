package keeper

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	interchainquerykeeper "github.com/Stride-Labs/stride/v24/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
)

type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	bankKeeper     types.BankKeeper
	transferKeeper types.TransferKeeper
	icqKeeper      interchainquerykeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	transferKeeper types.TransferKeeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		bankKeeper:     bankKeeper,
		transferKeeper: transferKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetLastUpdateTime stores the last time prices were updated
func (k Keeper) SetLastUpdateTime(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockTime().UnixNano()))
	store.Set([]byte(types.KeyLastUpdateTime), bz)
}

// GetLastUpdateTime retrieves the last time prices were updated
func (k Keeper) GetLastUpdateTime(ctx sdk.Context) (time.Time, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.KeyLastUpdateTime))
	if bz == nil {
		return time.Time{}, fmt.Errorf("last update time not found")
	}
	nanos := int64(binary.BigEndian.Uint64(bz))
	return time.Unix(0, nanos), nil
}

// SetTokenPrice stores price data for a token
func (k Keeper) SetTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice) error {
	store := ctx.KVStore(k.storeKey)
	key := types.TokenPriceKey(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	bz, err := k.cdc.Marshal(&tokenPrice)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// RemoveTokenPrice removes price data for a token
func (k Keeper) RemoveTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice) {
	store := ctx.KVStore(k.storeKey)
	key := types.TokenPriceKey(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	store.Delete(key)
}

func (k Keeper) SetTokenPriceQueryInProgress(ctx sdk.Context, tokenPrice types.TokenPrice, queryInProgress bool) error {
	tokenPrice, err := k.GetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return err
	}

	tokenPrice.QueryInProgress = queryInProgress
	err = k.SetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return err
	}

	return nil
}

// GetTokenPrice retrieves price data for a token
func (k Keeper) GetTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice) (types.TokenPrice, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.TokenPriceKey(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)

	bz := store.Get(key)
	if bz == nil {
		return types.TokenPrice{}, fmt.Errorf("price not found for baseDenom='%s' quoteDenom='%s' poolId='%s'", tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	}

	var price types.TokenPrice
	if err := k.cdc.Unmarshal(bz, &price); err != nil {
		return types.TokenPrice{}, err
	}

	return price, nil
}

// GetTokenPriceByDenom retrieves all price data for a base denom
func (k Keeper) GetTokenPricesByDenom(ctx sdk.Context, baseDenom string) (map[string]*types.TokenPrice, error) {
	store := ctx.KVStore(k.storeKey)

	// Create prefix iterator for all keys starting with baseDenom
	iterator := sdk.KVStorePrefixIterator(store, types.TokenPriceByDenomKey(baseDenom))
	defer iterator.Close()

	prices := make(map[string]*types.TokenPrice)

	for ; iterator.Valid(); iterator.Next() {
		var price types.TokenPrice
		if err := k.cdc.Unmarshal(iterator.Value(), &price); err != nil {
			return nil, err
		}

		// Use quoteDenom as the map key
		prices[price.QuoteDenom] = &price
	}

	return prices, nil
}

// GetAllTokenPrices retrieves all stored token prices
func (k Keeper) GetAllTokenPrices(ctx sdk.Context) []types.TokenPrice {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.KeyPricePrefix))
	defer iterator.Close()

	prices := []types.TokenPrice{}
	for ; iterator.Valid(); iterator.Next() {
		var price types.TokenPrice
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
	}

	return prices
}
