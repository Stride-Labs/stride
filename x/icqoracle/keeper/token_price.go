package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/utils"
	"github.com/Stride-Labs/stride/v28/x/icqoracle/types"
)

// SetTokenPrice stores price query for a token
func (k Keeper) SetTokenPrice(ctx sdk.Context, tokenPrice types.TokenPrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenPricePrefix)
	key := types.TokenPriceKey(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	bz := k.cdc.MustMarshal(&tokenPrice)
	store.Set(key, bz)
}

// RemoveTokenPrice removes price query for a token
func (k Keeper) RemoveTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenPricePrefix)
	key := types.TokenPriceKey(baseDenom, quoteDenom, osmosisPoolId)
	store.Delete(key)
}

// Updates the token price when a query is requested
func (k Keeper) SetQueryInProgress(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId uint64) error {
	tokenPrice, err := k.GetTokenPrice(ctx, baseDenom, quoteDenom, osmosisPoolId)
	if err != nil {
		return err
	}

	tokenPrice.QueryInProgress = true
	tokenPrice.LastRequestTime = ctx.BlockTime()

	k.SetTokenPrice(ctx, tokenPrice)

	return nil
}

// Updates the token price when a query response is received
func (k Keeper) SetQueryComplete(ctx sdk.Context, tokenPrice types.TokenPrice, newSpotPrice sdkmath.LegacyDec) {
	tokenPrice.SpotPrice = newSpotPrice
	tokenPrice.QueryInProgress = false
	tokenPrice.LastResponseTime = ctx.BlockTime()
	k.SetTokenPrice(ctx, tokenPrice)
}

// GetTokenPrice retrieves price data for a token
func (k Keeper) GetTokenPrice(ctx sdk.Context, baseDenom string, quoteDenom string, osmosisPoolId uint64) (types.TokenPrice, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenPricePrefix)
	key := types.TokenPriceKey(baseDenom, quoteDenom, osmosisPoolId)

	bz := store.Get(key)
	if bz == nil {
		return types.TokenPrice{}, fmt.Errorf("token price not found for baseDenom='%s' quoteDenom='%s' poolId='%d'", baseDenom, quoteDenom, osmosisPoolId)
	}

	var price types.TokenPrice
	if err := k.cdc.Unmarshal(bz, &price); err != nil {
		return types.TokenPrice{}, errorsmod.Wrapf(err, "unable to unmarshal token price")
	}

	return price, nil
}

// GetTokenPriceByDenom retrieves all price data for a base denom
// Returned as a mapping of each quote denom to the spot price
func (k Keeper) GetTokenPricesByDenom(ctx sdk.Context, baseDenom string) (map[string]*types.TokenPrice, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenPricePrefix)

	// Create prefix iterator for all keys starting with baseDenom
	iterator := storetypes.KVStorePrefixIterator(store, types.TokenPriceByDenomKey(baseDenom))
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
//   - All available prices with a common quote token are stale (exceeded the expiration timeout)
func (k Keeper) GetTokenPriceForQuoteDenom(ctx sdk.Context, baseDenom string, quoteDenom string) (sdkmath.LegacyDec, error) {
	// First attempt: Try to get the price with baseDenom as the base token and quoteDenom as the quote token
	price, errDirect := k.getTokenPriceForQuoteDenomImpl(ctx, baseDenom, quoteDenom)
	if errDirect == nil {
		return price, nil
	}

	// Second attempt: If the first attempt fails, try the reverse - use quoteDenom as the base token
	// and baseDenom as the quote token, then invert the price (1/price)
	price, errInverted := k.getTokenPriceForQuoteDenomImpl(ctx, quoteDenom, baseDenom)
	if errInverted == nil {
		// Invert the price to get the correct exchange rate
		price = sdkmath.LegacyNewDec(1).Quo(price)

		return price, nil
	}

	// If both attempts fail, return an error
	return sdkmath.LegacyDec{}, errorsmod.Wrapf(types.ErrQuotePriceNotFound,
		"no price found for baseDenom '%s' in terms of quoteDenom '%s' [%s], and no price found for '%s' in terms of '%s' [%s]",
		baseDenom, quoteDenom, errDirect, quoteDenom, baseDenom, errInverted)
}

// getTokenPriceForQuoteDenomImpl is the internal implementation that attempts to get the price
// for baseDenom in terms of quoteDenom by finding a common quote token. It returns an error
// if no valid price path can be found.
func (k Keeper) getTokenPriceForQuoteDenomImpl(ctx sdk.Context, baseDenom string, quoteDenom string) (price sdkmath.LegacyDec, err error) {
	// Get all price for baseToken
	baseTokenPrices, err := k.GetTokenPricesByDenom(ctx, baseDenom)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("error getting price for '%s': %w", baseDenom, err)
	}
	if len(baseTokenPrices) == 0 {
		return sdkmath.LegacyDec{}, fmt.Errorf("no price for baseDenom '%s'", baseDenom)
	}

	// Get price expiration timeout
	params := k.GetParams(ctx)
	priceExpirationTimeoutSec := utils.UintToInt(params.PriceExpirationTimeoutSec)

	// Check if baseDenom already has a price for quoteDenom
	foundAlreadyHasStalePrice := false
	foundHasUninitializedPrice := false
	if price, ok := baseTokenPrices[quoteDenom]; ok {
		if ctx.BlockTime().Unix()-price.LastResponseTime.Unix() <= priceExpirationTimeoutSec {
			if price.SpotPrice.IsZero() {
				foundHasUninitializedPrice = true
			} else {
				return price.SpotPrice, nil
			}
		} else {
			foundAlreadyHasStalePrice = true
		}
	}

	// Get all price for quoteToken
	quoteTokenPrices, err := k.GetTokenPricesByDenom(ctx, quoteDenom)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("error getting price for '%s': %w", quoteDenom, err)
	}
	if len(quoteTokenPrices) == 0 {
		return sdkmath.LegacyDec{}, fmt.Errorf("no price for quoteDenom '%s' (foundAlreadyHasStalePrice='%v', foundHasUninitializedPrice='%v')",
			quoteDenom, foundAlreadyHasStalePrice, foundHasUninitializedPrice)
	}

	// Init price
	price = sdkmath.LegacyZeroDec()

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
				if ctx.BlockTime().Unix()-baseTokenPrice.LastResponseTime.Unix() > priceExpirationTimeoutSec {
					foundBaseTokenStalePrice = true
					continue
				}
				if ctx.BlockTime().Unix()-quoteTokenPrice.LastResponseTime.Unix() > priceExpirationTimeoutSec {
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
		return sdkmath.LegacyDec{}, fmt.Errorf(
			"could not calculate price for baseToken='%s' quoteToken='%s' "+
				"(foundCommonQuoteToken='%v', foundBaseTokenStalePrice='%v', "+
				"foundQuoteTokenStalePrice='%v', foundQuoteTokenZeroPrice='%v', foundAlreadyHasStalePrice='%v')",
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
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.TokenPricePrefix)
	defer iterator.Close()

	prices := []types.TokenPrice{}
	for ; iterator.Valid(); iterator.Next() {
		var price types.TokenPrice
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
	}

	return prices
}
