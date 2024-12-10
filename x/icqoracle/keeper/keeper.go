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

// UpdatePrices queries and updates all token prices
func (k Keeper) UpdatePrices(ctx sdk.Context) error {
	// Query prices from Osmosis via ICQ
	prices, err := k.QueryOsmosisPrices(ctx)
	if err != nil {
		return err
	}

	// Update stored prices
	for _, price := range prices {
		if err := k.SetTokenPrice(ctx, price); err != nil {
			return err
		}
	}

	return nil
}

// SetTokenPrice stores price data for a token
func (k Keeper) SetTokenPrice(ctx sdk.Context, price types.TokenPrice) error {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s/%s/%s", types.KeyPricePrefix, price.BaseDenom, price.QuoteDenom))

	bz, err := k.cdc.Marshal(&price)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// RemoveTokenPrice removes price data for a token
func (k Keeper) RemoveTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s/%s/%s", types.KeyPricePrefix, baseDenom, quoteDenom))
	store.Delete(key)
}

// GetTokenPrice retrieves price data for a token
func (k Keeper) GetTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string) (types.TokenPrice, error) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("%s/%s/%s", types.KeyPricePrefix, baseDenom, quoteDenom))

	bz := store.Get(key)
	if bz == nil {
		return types.TokenPrice{}, fmt.Errorf("price not found for base denom '%s' & quote denom '%s'", baseDenom, quoteDenom)
	}

	var price types.TokenPrice
	if err := k.cdc.Unmarshal(bz, &price); err != nil {
		return types.TokenPrice{}, err
	}

	return price, nil
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

// QueryOsmosisPrices implements the ICQ price query
func (k Keeper) QueryOsmosisPrices(ctx sdk.Context) ([]types.TokenPrice, error) {
	// TODO: Implement actual ICQ query to Osmosis
	// This is a placeholder that returns mock data
	mockPrices := []types.TokenPrice{
		{
			BaseDenom:  "uatom",
			QuoteDenom: "ustrd",
			Price:      sdk.NewDec(10),
			UpdatedAt:  ctx.BlockTime(),
		},
		{
			BaseDenom:  "uosmo",
			QuoteDenom: "ustrd",
			Price:      sdk.NewDec(5),
			UpdatedAt:  ctx.BlockTime(),
		},
	}
	return mockPrices, nil
}
