package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
)

func (s *KeeperTestSuite) TestQueryTokenPrice() {
	// Create token price entry
	baseDenom := "uatom"
	quoteDenom := "uusdc"
	poolId := "1"
	expectedPrice := sdkmath.LegacyNewDec(1000000)

	tokenPrice := types.TokenPrice{
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: poolId,
		SpotPrice:     expectedPrice,
	}
	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err, "no error expected when setting token price %+v", tokenPrice)

	// Query for the token price
	req := &types.QueryTokenPriceRequest{
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
		PoolId:     poolId,
	}
	resp, err := s.App.ICQOracleKeeper.TokenPrice(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying token price")
	s.Require().Equal(expectedPrice, resp.TokenPrice.SpotPrice, "token price")

	// Query with invalid request
	_, err = s.App.ICQOracleKeeper.TokenPrice(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")
}

func (s *KeeperTestSuite) TestQueryTokenPrices() {
	// Create multiple token prices
	expectedPrices := []types.TokenPrice{
		{
			BaseDenom:     "uatom",
			QuoteDenom:    "uusdc",
			OsmosisPoolId: "1",
			SpotPrice:     sdkmath.LegacyNewDec(1000000),
		},
		{
			BaseDenom:     "uosmo",
			QuoteDenom:    "uusdc",
			OsmosisPoolId: "2",
			SpotPrice:     sdkmath.LegacyNewDec(2000000),
		},
	}

	for _, price := range expectedPrices {
		err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, price)
		s.Require().NoError(err, "no error expected when setting token price %+v", price)
	}

	// Query all token prices
	req := &types.QueryTokenPricesRequest{}
	resp, err := s.App.ICQOracleKeeper.TokenPrices(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all token prices")
	s.Require().Equal(expectedPrices, resp.TokenPrices, "token prices")

	// Query with invalid request
	_, err = s.App.ICQOracleKeeper.TokenPrices(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")
}

func (s *KeeperTestSuite) TestQueryParams() {
	// Set parameters
	expectedParams := types.Params{
		OsmosisChainId:            "osmosis-1",
		OsmosisConnectionId:       "connection-2",
		UpdateIntervalSec:         5 * 60,  // 5 min
		PriceExpirationTimeoutSec: 10 * 60, // 10 min
		IcqTimeoutSec:             2 * 60,  // 2 min
	}
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, expectedParams)
	s.Require().NoError(err, "no error expected when setting params")

	// Query parameters
	req := &types.QueryParamsRequest{}
	resp, err := s.App.ICQOracleKeeper.Params(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying params")
	s.Require().Equal(expectedParams, resp.Params, "params")

	// Query with invalid request
	_, err = s.App.ICQOracleKeeper.Params(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")
}

func (s *KeeperTestSuite) TestQueryTokenPriceForQuoteDenomSimple() {
	// Setup params
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, types.Params{
		PriceExpirationTimeoutSec: 60,
	})
	s.Require().NoError(err, "no error expected when setting params")

	// Create token price with same quote denom
	baseDenom := "uatom"
	quoteDenom := "uusdc"
	expectedPrice := sdkmath.LegacyNewDec(1000000)

	tokenPrice := types.TokenPrice{
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: "1",
		SpotPrice:     expectedPrice,
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err, "no error expected when setting token price %+v", tokenPrice)

	// Query for token price using quote denom
	req := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
	resp, err := s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying token price for quote denom")
	s.Require().Equal(expectedPrice, resp.Price, "token price")

	// Query with invalid request
	_, err = s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")

	// Query with non-existent denom pair
	reqNonExistent := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  "nonexistent",
		QuoteDenom: "nonexistent",
	}
	_, err = s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), reqNonExistent)
	s.Require().Error(err, "error expected when querying non-existent denom pair")
}

func (s *KeeperTestSuite) TestQueryTokenPriceForQuoteDenom() {
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, types.Params{
		PriceExpirationTimeoutSec: 60,
	})
	s.Require().NoError(err, "no error expected when setting params")

	// Create two token prices with same quote denom
	baseDenom1 := "uatom"
	baseDenom2 := "uosmo"
	quoteDenom := "uusdc"
	expectedPrice1 := sdkmath.LegacyNewDec(1000000)
	expectedPrice2 := sdkmath.LegacyNewDec(2000000)

	// Set uatom price
	tokenPrice1 := types.TokenPrice{
		BaseDenom:     baseDenom1,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: "1",
		SpotPrice:     expectedPrice1,
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice1)
	s.Require().NoError(err, "no error expected when setting token price %+v", tokenPrice1)

	// Set uosmo price
	tokenPrice2 := types.TokenPrice{
		BaseDenom:     baseDenom2,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: "2",
		SpotPrice:     expectedPrice2,
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice2)
	s.Require().NoError(err, "no error expected when setting token price %+v", tokenPrice2)

	// Query for token price using a common quote denom
	req := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  baseDenom1,
		QuoteDenom: baseDenom2,
	}
	resp, err := s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying token price for quote denom")
	s.Require().Equal(expectedPrice1.Quo(expectedPrice2), resp.Price, "token price")
}

func (s *KeeperTestSuite) TestQueryTokenPriceForQuoteDenomStalePrice() {
	// Set up parameters with short expiration time
	params := types.Params{
		PriceExpirationTimeoutSec: 60, // 1 minute timeout
	}
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Create token prices
	baseDenom := "uatom"
	quoteDenom := "uusdc"
	expectedPrice := sdkmath.LegacyNewDec(1000000)

	tokenPrice := types.TokenPrice{
		BaseDenom:     baseDenom,
		QuoteDenom:    quoteDenom,
		OsmosisPoolId: "1",
		SpotPrice:     expectedPrice,
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Fast forward block time to make price stale
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute * 2))

	// Query should fail due to stale price
	req := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
	_, err = s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err, "expected error for stale price")
	s.Require().Contains(err.Error(), "foundAlreadyHasStalePrice='true'", "error should indicate price calculation failure")
}

func (s *KeeperTestSuite) TestQueryTokenPriceForQuoteDenomZeroPrice() {
	// Create token prices with zero price for quote token
	baseDenom := "uatom"
	quoteDenom := "uusdc"
	intermediateQuote := "uosmo"

	// Set base token price
	tokenPrice1 := types.TokenPrice{
		BaseDenom:     baseDenom,
		QuoteDenom:    intermediateQuote,
		OsmosisPoolId: "1",
		SpotPrice:     sdkmath.LegacyNewDec(1000000),
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice1)
	s.Require().NoError(err)

	// Set quote token price with zero value
	tokenPrice2 := types.TokenPrice{
		BaseDenom:     quoteDenom,
		QuoteDenom:    intermediateQuote,
		OsmosisPoolId: "2",
		SpotPrice:     sdkmath.LegacyZeroDec(),
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice2)
	s.Require().NoError(err)

	// Query should fail due to zero price
	req := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
	_, err = s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err, "expected error for zero price")
	s.Require().Contains(err.Error(), "could not calculate price", "error should indicate price calculation failure")
}

func (s *KeeperTestSuite) TestQueryTokenPriceForQuoteDenomNoCommonQuote() {
	// Create token prices with different quote denoms
	baseDenom := "uatom"
	quoteDenom := "uusdc"

	// Set base token price with one quote denom
	tokenPrice1 := types.TokenPrice{
		BaseDenom:     baseDenom,
		QuoteDenom:    "quote1",
		OsmosisPoolId: "1",
		SpotPrice:     sdkmath.LegacyNewDec(1000000),
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice1)
	s.Require().NoError(err)

	// Set quote token price with different quote denom
	tokenPrice2 := types.TokenPrice{
		BaseDenom:     quoteDenom,
		QuoteDenom:    "quote2",
		OsmosisPoolId: "2",
		SpotPrice:     sdkmath.LegacyNewDec(2000000),
		UpdatedAt:     s.Ctx.BlockTime(),
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice2)
	s.Require().NoError(err)

	// Query should fail due to no common quote denom
	req := &types.QueryTokenPriceForQuoteDenomRequest{
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
	}
	_, err = s.App.ICQOracleKeeper.TokenPriceForQuoteDenom(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().Error(err, "expected error when no common quote denom exists")
	s.Require().Contains(err.Error(), "could not calculate price", "error should indicate price calculation failure")
}
