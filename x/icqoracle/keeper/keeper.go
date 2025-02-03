package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

type Keeper struct {
	cdc               codec.Codec
	storeKey          storetypes.StoreKey
	IcqKeeper         types.IcqKeeper
	ibcTransferKeeper types.IbcTransferKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	icqKeeper types.IcqKeeper,
	ibcTransferKeeper types.IbcTransferKeeper,
) *Keeper {
	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		IcqKeeper:         icqKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetTokenPrice stores price query for a token
func (k Keeper) SetTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PriceQueryPrefix)
	key := types.TokenPriceQueryKey(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)

	bz, err := k.cdc.Marshal(&tokenPrice)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// RemoveTokenPrice removes price query for a token
func (k Keeper) RemoveTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PriceQueryPrefix)
	key := types.TokenPriceQueryKey(baseDenom, quoteDenom, osmosisPoolId)
	store.Delete(key)
}

func (k Keeper) SetTokenPriceQueryInProgress(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId string, queryInProgress bool) error {
	tokenPrice, err := k.GetTokenPrice(ctx, baseDenom, quoteDenom, osmosisPoolId)
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
func (k Keeper) GetTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId string) (types.TokenPrice, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PriceQueryPrefix)
	key := types.TokenPriceQueryKey(baseDenom, quoteDenom, osmosisPoolId)

	bz := store.Get(key)
	if bz == nil {
		return types.TokenPrice{}, fmt.Errorf("price not found for baseDenom='%s' quoteDenom='%s' poolId='%s'", baseDenom, quoteDenom, osmosisPoolId)
	}

	var price types.TokenPrice
	if err := k.cdc.Unmarshal(bz, &price); err != nil {
		return types.TokenPrice{}, err
	}

	return price, nil
}

// GetTokenPriceByDenom retrieves all price data for a base denom
// Returned as a mapping of each quote denom to the spot price
func (k Keeper) GetTokenPricesByDenom(ctx sdk.Context, baseDenom string) (map[string]*types.TokenPrice, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PriceQueryPrefix)

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

// GetTokenPriceForQuoteDenom calculates and retrieves the exchange rate between two tokens.
// The exchange rate is determined by finding a common quote token between both tokens,
// and then dividing their respective spot prices.
//
// For example, if we have:
//   - baseToken/USDC = 10
//   - quoteToken/USDC = 5
//
// Then:
//   - baseToken/quoteToken = 10/5 = 2
//
// Parameters:
//   - ctx: SDK Context for accessing the store
//   - baseDenom: The denom of the token to get the price for
//   - quoteDenom: The denom to price the base token in
//
// Returns:
//   - math.LegacyDec: The exchange rate of 1 baseToken in terms of quoteToken
//   - error: Returns an error if:
//   - No prices exist for either token
//   - No common quote token exists between the two tokens
//   - All available prices with a common quote token are stale (exceeded the stale timeout)
func (k Keeper) GetTokenPriceForQuoteDenom(ctx sdk.Context, baseDenom string, quoteDenom string) (price math.LegacyDec, err error) {
	// Get all price for baseToken
	baseTokenPrices, err := k.GetTokenPricesByDenom(ctx, baseDenom)
	if err != nil {
		return math.LegacyDec{}, fmt.Errorf("error getting price for '%s': %w", baseDenom, err)
	}
	if len(baseTokenPrices) == 0 {
		return math.LegacyDec{}, fmt.Errorf("no price for baseDenom '%s'", baseDenom)
	}

	// Get price expiration timeout
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, fmt.Errorf("error getting params: %w", err)
	}
	priceExpirationTimeoutSec := int64(params.PriceExpirationTimeoutSec)

	// Check if baseDenom already has a price for quoteDenom
	foundAlreadyHasStalePrice := false
	if price, ok := baseTokenPrices[quoteDenom]; ok {
		if ctx.BlockTime().Unix()-price.UpdatedAt.Unix() <= priceExpirationTimeoutSec {
			return price.SpotPrice, nil
		} else {
			foundAlreadyHasStalePrice = true
		}
	}

	// Get all price for quoteToken
	quoteTokenPrices, err := k.GetTokenPricesByDenom(ctx, quoteDenom)
	if err != nil {
		return math.LegacyDec{}, fmt.Errorf("error getting price for '%s': %w", quoteDenom, err)
	}
	if len(quoteTokenPrices) == 0 {
		return math.LegacyDec{}, fmt.Errorf("no price for quoteDenom '%s' (foundAlreadyHasStalePrice='%v')", quoteDenom, foundAlreadyHasStalePrice)
	}

	// Init price
	price = math.LegacyZeroDec()

	// Define flags to allow for better error messages
	foundCommonQuoteToken := false
	foundBaseTokenStalePrice := false
	foundQuoteTokenStalePrice := false
	foundQuoteTokenZeroPrice := false

	// Find a common quote denom and calculate baseToken to quoteToken price
	for commonQuoteDenom1, baseTokenPrice := range baseTokenPrices {
		for commonQuoteDenom2, quoteTokenPrice := range quoteTokenPrices {
			if commonQuoteDenom1 == commonQuoteDenom2 {
				foundCommonQuoteToken = true

				// Check that both prices are not stale
				if ctx.BlockTime().Unix()-baseTokenPrice.UpdatedAt.Unix() > priceExpirationTimeoutSec {
					foundBaseTokenStalePrice = true
					continue
				}
				if ctx.BlockTime().Unix()-quoteTokenPrice.UpdatedAt.Unix() > priceExpirationTimeoutSec {
					foundQuoteTokenStalePrice = true
					continue
				}

				// Check that quote price is not zero to prevent division by zero
				if quoteTokenPrice.SpotPrice.IsZero() {
					foundQuoteTokenZeroPrice = true
					continue
				}

				// Calculate the price of 1 baseToken in quoteToken
				price = baseTokenPrice.SpotPrice.Quo(quoteTokenPrice.SpotPrice)
				break
			}
		}
	}

	if price.IsZero() {
		return math.LegacyDec{}, fmt.Errorf(
			"could not calculate price for baseToken='%s' quoteToken='%s' (foundCommonQuoteToken='%v', foundBaseTokenStalePrice='%v', foundQuoteTokenStalePrice='%v', foundQuoteTokenZeroPrice='%v', foundAlreadyHasStalePrice='%v')",
			baseDenom,
			quoteDenom,
			foundCommonQuoteToken,
			foundBaseTokenStalePrice,
			foundQuoteTokenStalePrice,
			foundQuoteTokenZeroPrice,
			foundAlreadyHasStalePrice,
		)
	}

	return price, nil
}

// GetAllTokenPrices retrieves all stored token prices
func (k Keeper) GetAllTokenPrices(ctx sdk.Context) []types.TokenPrice {
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte(types.PriceQueryPrefix))
	defer iterator.Close()

	prices := []types.TokenPrice{}
	for ; iterator.Valid(); iterator.Next() {
		var price types.TokenPrice
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
	}

	return prices
}

// GetStoreKey returns the store key
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}
