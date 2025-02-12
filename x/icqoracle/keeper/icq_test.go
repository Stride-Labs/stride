package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"
	"github.com/osmosis-labs/osmosis/osmomath"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/keeper"
	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

// Mock ICQ Keeper struct
type MockICQKeeper struct {
	SubmitICQRequestFn func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error
}

func (m MockICQKeeper) SubmitICQRequest(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
	if m.SubmitICQRequestFn != nil {
		return m.SubmitICQRequestFn(ctx, query, forceUnique)
	}
	return nil
}

func (s *KeeperTestSuite) TestSubmitOsmosisPoolICQ_Success() {
	var submittedQuery icqtypes.Query

	// Setup mock ICQ keeper with custom behavior
	s.mockICQKeeper = MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			submittedQuery = query
			return nil
		},
	}
	s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper

	// Set up test parameters
	tokenPrice := types.TokenPrice{
		BaseDenom:     "uatom",
		QuoteDenom:    "uusdc",
		OsmosisPoolId: 1,
	}

	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	params := types.Params{
		OsmosisChainId:      "osmosis-1",
		OsmosisConnectionId: "connection-0",
		UpdateIntervalSec:   60,
	}
	s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

	// Verify tokenPrice.QueryInProgress before
	s.Require().False(tokenPrice.QueryInProgress)

	// Submit ICQ request
	err := s.App.ICQOracleKeeper.SubmitOsmosisPoolICQ(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Verify the captured query data
	s.Require().Equal(params.OsmosisChainId, submittedQuery.ChainId)
	s.Require().Equal(params.OsmosisConnectionId, submittedQuery.ConnectionId)
	s.Require().Equal(icqtypes.OSMOSIS_TWAP_STORE_QUERY_WITH_PROOF, submittedQuery.QueryType)
	s.Require().Equal(types.ModuleName, submittedQuery.CallbackModule)
	s.Require().Equal(keeper.ICQCallbackID_OsmosisPool, submittedQuery.CallbackId)

	// Verify tokenPrice.QueryInProgress after
	tokenPriceAfter, err := s.App.ICQOracleKeeper.GetTokenPrice(
		s.Ctx,
		tokenPrice.BaseDenom,
		tokenPrice.QuoteDenom,
		tokenPrice.OsmosisPoolId,
	)
	s.Require().NoError(err)

	s.Require().True(tokenPriceAfter.QueryInProgress, "query in progress")
	s.Require().Equal(tokenPriceAfter.LastRequestTime, s.Ctx.BlockTime(), "query request time")

	// Verify callback data contains the token price
	var decodedTokenPrice types.TokenPrice
	err = s.App.AppCodec().Unmarshal(submittedQuery.CallbackData, &decodedTokenPrice)
	s.Require().NoError(err)
	s.Require().Equal(tokenPrice.BaseDenom, decodedTokenPrice.BaseDenom)
	s.Require().Equal(tokenPrice.QuoteDenom, decodedTokenPrice.QuoteDenom)
	s.Require().Equal(tokenPrice.OsmosisPoolId, decodedTokenPrice.OsmosisPoolId)

	// Verify timeout settings
	expectedTimeout := time.Duration(params.UpdateIntervalSec) * time.Second
	s.Require().Equal(expectedTimeout, submittedQuery.TimeoutDuration)
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, submittedQuery.TimeoutPolicy)
}

func (s *KeeperTestSuite) TestSubmitOsmosisPoolICQ_Errors() {
	testCases := []struct {
		name          string
		setup         func()
		tokenPrice    types.TokenPrice
		expectedError string
	}{
		{
			name: "token price not found",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

				// Token price will not be added to the store in this test
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: 1,
			},
			expectedError: "token price not found",
		},
		{
			name: "error submitting ICQ request",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

				// Mock ICQ keeper to return error
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
						return fmt.Errorf("mock ICQ submit error")
					},
				}
				s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: 1,
			},
			expectedError: "Error submitting OsmosisPool ICQ",
		},
		{
			name: "error setting query in progress",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

				// Setup mock ICQ keeper with success response
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
						// Remove token price so set query in progress will fail to get it after SubmitICQRequest returns
						s.App.ICQOracleKeeper.RemoveTokenPrice(ctx, "uatom", "uusdc", 1)
						return nil
					},
				}
				s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper

				// Don't set the token price first, which will cause SetTokenPriceQueryInProgress to fail
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: 1,
			},
			expectedError: "Error updating token price query to in progress",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset context for each test case
			s.SetupTest()

			// Run test case setup
			tc.setup()

			// If this is not an error case testing query in progress,
			// set the token price first
			if tc.expectedError != "token price not found" {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tc.tokenPrice)
			}

			// Execute
			err := s.App.ICQOracleKeeper.SubmitOsmosisPoolICQ(s.Ctx, tc.tokenPrice)

			// Verify results
			if tc.expectedError != "" {
				s.Require().ErrorContains(err, tc.expectedError)
			} else {
				s.Require().NoError(err)

				// For successful case, verify query in progress was set
				tokenPriceAfter, err := s.App.ICQOracleKeeper.GetTokenPrice(
					s.Ctx,
					tc.tokenPrice.BaseDenom,
					tc.tokenPrice.QuoteDenom,
					tc.tokenPrice.OsmosisPoolId,
				)
				s.Require().NoError(err)
				s.Require().True(tokenPriceAfter.QueryInProgress)
			}
		})
	}
}

// Helper function to create mock twap data
// P0 and P1 store the relative ratio of assets in the pool
//
//	P0 is ratio of Asset0 / Asset1
//	P1 is ratio of Asset1 / Asset0
//
// We want to ratio of quote denom to base denom, which will give the price of base denom
// in terms of quote denom
//
// For this function, we'll always return a price of 1.5 for the baseDenom in terms of quote denom
// However, the assets may be inverted depending on the parameters
func (s *KeeperTestSuite) createMockTwapData(baseDenom, quoteDenom, assetDenom0, assetDenom1 string) []byte {
	baseAssetPrice := sdk.MustNewDecFromStr("1.5")
	quoteAssetPrice := sdk.OneDec().Quo(baseAssetPrice)

	pool := types.OsmosisTwapRecord{
		Asset0Denom: assetDenom0,
		Asset1Denom: assetDenom1,
	}

	// If asset0 is the quote denom, then we want P0 to give us the price of base asset
	if assetDenom0 == quoteDenom {
		s.Require().Equal(assetDenom1, baseDenom, "Invalid test case setup, baseDenom not asset 1")

		pool.P0LastSpotPrice = baseAssetPrice // <- price we use
		pool.P1LastSpotPrice = quoteAssetPrice
	} else {
		s.Require().Equal(assetDenom0, baseDenom, "Invalid test case setup, baseDenom not asset 0")
		s.Require().Equal(assetDenom1, quoteDenom, "Invalid test case setup, quoteDenom not asset 1")

		// If asset0 is the base denom, then we want P1 to give us the price of the base asset
		pool.P0LastSpotPrice = quoteAssetPrice
		pool.P1LastSpotPrice = baseAssetPrice // <- price we use
	}

	bz, err := proto.Marshal(&pool)
	s.Require().NoError(err, "no error expected when marshaling mock pool data")
	return bz
}

func (s *KeeperTestSuite) TestOsmosisPoolCallback() {
	// Setup base test parameters used across test cases
	baseTokenPrice := types.TokenPrice{
		BaseDenom:         "uatom",
		QuoteDenom:        "uusdc",
		OsmosisPoolId:     1,
		OsmosisBaseDenom:  "ibc/uatom",
		OsmosisQuoteDenom: "ibc/uusdc",
		SpotPrice:         math.LegacyNewDec(2),
		QueryInProgress:   true,
	}

	testCases := []struct {
		name          string
		setup         func() ([]byte, icqtypes.Query)
		expectedError string
		verify        func(err error)
	}{
		{
			name: "invalid callback data",
			setup: func() ([]byte, icqtypes.Query) {
				return []byte{}, icqtypes.Query{
					CallbackData: []byte("invalid callback data"),
				}
			},
			expectedError: "Error deserializing query.CallbackData",
		},
		{
			name: "token price not found",
			setup: func() ([]byte, icqtypes.Query) {
				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Don't set the token price in state
				return []byte{}, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "price not found",
		},
		{
			name: "query not in progress",
			setup: func() ([]byte, icqtypes.Query) {
				tokenPrice := baseTokenPrice
				tokenPrice.QueryInProgress = false
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&tokenPrice)
				s.Require().NoError(err)

				return []byte{}, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "",
			verify: func(err error) {
				s.Require().NoError(err, "no error expected when query not in progress")
			},
		},
		{
			name: "invalid pool data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				return []byte("invalid pool data"), icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "Error determining spot price from query response",
		},
		{
			name: "successful update with valid pool data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create mock pool data with expected price of 2.0
				poolData := s.createMockTwapData(
					baseTokenPrice.OsmosisBaseDenom,
					baseTokenPrice.OsmosisQuoteDenom,
				)

				return poolData, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "",
			verify: func(err error) {
				s.Require().NoError(err)

				// Verify token price was updated correctly
				tokenPrice := s.MustGetTokenPrice(
					baseTokenPrice.BaseDenom,
					baseTokenPrice.QuoteDenom,
					baseTokenPrice.OsmosisPoolId,
				)

				// Verify updated fields
				s.Require().False(tokenPrice.QueryInProgress)
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.LastResponseTime.UnixNano())
				s.Require().InDelta(1.5, tokenPrice.SpotPrice.MustFloat64(), 0.01)
			},
		},
		{
			name: "successful update with inverse pool data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with swapped denoms
				poolData := s.createMockTwapData(
					baseTokenPrice.OsmosisQuoteDenom, // Swapped!
					baseTokenPrice.OsmosisBaseDenom,  // Swapped!
				)

				return poolData, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "",
			verify: func(err error) {
				s.Require().NoError(err)

				// Verify token price was updated correctly
				tokenPrice := s.MustGetTokenPrice(
					baseTokenPrice.BaseDenom,
					baseTokenPrice.QuoteDenom,
					baseTokenPrice.OsmosisPoolId,
				)

				// Verify updated fields
				s.Require().False(tokenPrice.QueryInProgress)
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.LastResponseTime.UnixNano())
				s.Require().InDelta(1/1.5, tokenPrice.SpotPrice.MustFloat64(), 0.01) // inversed price
			},
		},
		{
			name: "empty query callback data",
			setup: func() ([]byte, icqtypes.Query) {
				return []byte{}, icqtypes.Query{
					CallbackData: []byte{},
				}
			},
			expectedError: "price not found for baseDenom='' quoteDenom='' poolId='0'",
		},
		{
			name: "nil query callback data",
			setup: func() ([]byte, icqtypes.Query) {
				return []byte{}, icqtypes.Query{
					CallbackData: nil,
				}
			},
			expectedError: "price not found for baseDenom='' quoteDenom='' poolId='0'",
		},
		{
			name: "corrupted token price in callback data",
			setup: func() ([]byte, icqtypes.Query) {
				// Create corrupted protobuf data
				corruptedData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
				return []byte{}, icqtypes.Query{
					CallbackData: corruptedData,
				}
			},
			expectedError: "Error deserializing query.CallbackData",
		},
		{
			name: "empty pool response data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				return []byte{}, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "Error determining spot price from query response",
		},
		{
			name: "nil pool response data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				return nil, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "Error determining spot price from query response",
		},
		{
			name: "corrupted pool response data",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create corrupted protobuf data
				corruptedPoolData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
				return corruptedPoolData, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "Error determining spot price from query response",
		},
		{
			name: "pool with empty tokens",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with empty token denoms
				pool := cltypes.OsmosisConcentratedLiquidityPool{
					Token0:           "",
					Token1:           "",
					CurrentSqrtPrice: osmomath.NewBigDecWithPrec(1224744871, 9),
				}
				poolBz, err := proto.Marshal(&pool)
				s.Require().NoError(err)

				return poolBz, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "Error determining spot price from query response",
		},
		{
			name: "pool with invalid sqrt price",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with invalid sqrt price
				pool := cltypes.OsmosisConcentratedLiquidityPool{
					Token0:           baseTokenPrice.OsmosisBaseDenom,
					Token1:           baseTokenPrice.OsmosisQuoteDenom,
					CurrentSqrtPrice: osmomath.NewBigDec(0), // Invalid sqrt price of 0
				}
				poolBz, err := proto.Marshal(&pool)
				s.Require().NoError(err)

				return poolBz, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "zero sqrt price",
		},
		{
			name: "pool with invalid inverse sqrt price",
			setup: func() ([]byte, icqtypes.Query) {
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with invalid sqrt price
				pool := cltypes.OsmosisConcentratedLiquidityPool{
					Token0:           baseTokenPrice.OsmosisQuoteDenom,
					Token1:           baseTokenPrice.OsmosisBaseDenom,
					CurrentSqrtPrice: osmomath.NewBigDec(0), // Invalid sqrt price of 0
				}
				poolBz, err := proto.Marshal(&pool)
				s.Require().NoError(err)

				return poolBz, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: "zero sqrt price",
		},
		{
			name: "error setting updated token price",
			setup: func() ([]byte, icqtypes.Query) {
				// First set the token price
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create valid pool data
				poolData := s.createMockTwapData(
					baseTokenPrice.OsmosisBaseDenom,
					baseTokenPrice.OsmosisQuoteDenom,
				)

				// Remove the token price from state right before setting the updated price
				s.App.ICQOracleKeeper.RemoveTokenPrice(s.Ctx, baseTokenPrice.BaseDenom, baseTokenPrice.QuoteDenom, baseTokenPrice.OsmosisPoolId)

				return poolData, icqtypes.Query{
					CallbackData: tokenPriceBz,
				}
			},
			expectedError: fmt.Sprintf(
				"price not found for baseDenom='%s' quoteDenom='%s' poolId='%d'",
				baseTokenPrice.BaseDenom,
				baseTokenPrice.QuoteDenom,
				baseTokenPrice.OsmosisPoolId),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset context for each test case
			s.SetupTest()

			// Run test case setup
			poolData, query := tc.setup()

			// Execute callback
			err := keeper.OsmosisPoolCallback(s.App.ICQOracleKeeper, s.Ctx, poolData, query)

			// Verify results
			if tc.expectedError != "" {
				s.Require().ErrorContains(err, tc.expectedError)
			} else if tc.verify != nil {
				tc.verify(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnmarshalSpotPriceFromOsmosis() {
	testCases := []struct {
		name          string
		tokenPrice    types.TokenPrice
		twapData      []byte
		expectedPrice math.LegacyDec
		expectedError string
	}{
		{
			name: "invalid pool data",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  6,
				QuoteDenomDecimals: 6,
			},
			twapData:      []byte("invalid pool data"),
			expectedError: "unable to unmarshal the query response",
		},
		{
			name: "successful price calculation with equal decimals",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  6,
				QuoteDenomDecimals: 6,
			},
			twapData: s.createMockTwapData(
				"ibc/atom", // base
				"ibc/usdc", // quote
				"ibc/atom", // asset0
				"ibc/usdc", // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5"), // 1.5 from mock pool data
		},
		{
			name: "successful price calculation with equal decimals and assets inverted",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  6,
				QuoteDenomDecimals: 6,
			},
			twapData: s.createMockTwapData(
				"ibc/atom", // base
				"ibc/usdc", // quote
				"ibc/usdc", // asset0
				"ibc/atom", // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5"), // 1.5 from mock pool data
		},
		{
			name: "successful price calculation with more base decimals",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/satoshi",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  8, // BTC has 8 decimals
				QuoteDenomDecimals: 6, // USDC has 6 decimals
			},
			twapData: s.createMockTwapData(
				"ibc/satoshi", // base
				"ibc/usdc",    // quote
				"ibc/satoshi", // asset0
				"ibc/usdc",    // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5").Mul(math.LegacyNewDec(100)), // 1.5 * 10^2
		},
		{
			name: "successful price calculation with more base decimals and assets inverted",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/satoshi",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  8, // BTC has 8 decimals
				QuoteDenomDecimals: 6, // USDC has 6 decimals
			},
			twapData: s.createMockTwapData(
				"ibc/satoshi", // base
				"ibc/usdc",    // quote
				"ibc/usdc",    // asset0
				"ibc/satoshi", // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5").Mul(math.LegacyNewDec(100)), // 1.5 * 10^2
		},
		{
			name: "successful price calculation with more quote decimals",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  6, // ATOM has 6 decimals
				QuoteDenomDecimals: 8, // Quote has 8 decimals
			},
			twapData: s.createMockTwapData(
				"ibc/atom", // base
				"ibc/usdc", // quote
				"ibc/atom", // asset0
				"ibc/usdc", // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5").Quo(math.LegacyNewDec(100)), // 1.5 / 10^2
		},
		{
			name: "successful price calculation with more quote decimals and assets inverted",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "ibc/usdc",
				BaseDenomDecimals:  6, // ATOM has 6 decimals
				QuoteDenomDecimals: 8, // Quote has 8 decimals
			},
			twapData: s.createMockTwapData(
				"ibc/atom", // base
				"ibc/usdc", // quote
				"ibc/usdc", // asset0
				"ibc/atom", // asset1
			),
			expectedPrice: sdk.MustNewDecFromStr("1.5").Quo(math.LegacyNewDec(100)), // 1.5 / 10^2
		},
		{
			name: "different denom ordering in pool",
			tokenPrice: types.TokenPrice{
				OsmosisBaseDenom:   "ibc/atom",
				OsmosisQuoteDenom:  "different_denom",
				BaseDenomDecimals:  6,
				QuoteDenomDecimals: 6,
			},
			twapData: s.createMockTwapData(
				"ibc/atom",
				"ibc/usdc",
				"ibc/atom",
				"ibc/usdc",
			),
			expectedError: "Assets in query response (ibc/atom, ibc/usdc) do not match denom's from token price (ibc/atom, different_denom)",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			spotPrice, err := keeper.UnmarshalSpotPriceFromOsmosis(s.App.ICQOracleKeeper, tc.tokenPrice, tc.twapData)

			if tc.expectedError != "" {
				s.Require().ErrorContains(err, tc.expectedError)
			} else {
				s.Require().NoError(err)
				s.Require().InDelta(
					tc.expectedPrice.MustFloat64(),
					spotPrice.MustFloat64(),
					0.000001,
					"expected price %v, got %v", tc.expectedPrice, spotPrice,
				)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAssertTwapAssetsMatchTokenPrice() {
	testCases := []struct {
		name          string
		twapRecord    types.OsmosisTwapRecord
		tokenPrice    types.TokenPrice
		expectedMatch bool
	}{
		{
			name:          "successful match - 1",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-a", OsmosisQuoteDenom: "denom-b"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 2",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-b", Asset1Denom: "denom-a"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-b", OsmosisQuoteDenom: "denom-a"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 3",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-b", OsmosisQuoteDenom: "denom-a"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 4",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-b", Asset1Denom: "denom-a"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-a", OsmosisQuoteDenom: "denom-b"},
			expectedMatch: true,
		},
		{
			name:          "mismatch osmo asset 0",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-z", Asset1Denom: "denom-b"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-a", OsmosisQuoteDenom: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch osmo asset 1",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-z"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-a", OsmosisQuoteDenom: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch route reward denom",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-z", OsmosisQuoteDenom: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch route host denom",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tokenPrice:    types.TokenPrice{OsmosisBaseDenom: "denom-a", OsmosisQuoteDenom: "denom-z"},
			expectedMatch: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := keeper.AssertTwapAssetsMatchTokenPrice(tc.twapRecord, tc.tokenPrice)
			if tc.expectedMatch {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestAdjustSpotPriceForDecimals() {
	testCases := []struct {
		name          string
		rawPrice      math.LegacyDec
		baseDecimals  int64
		quoteDecimals int64
		expectedPrice math.LegacyDec
	}{
		{
			name:          "equal decimals",
			rawPrice:      math.LegacyNewDec(1000),
			baseDecimals:  6,
			quoteDecimals: 6,
			expectedPrice: math.LegacyNewDec(1000),
		},
		{
			name:          "base has more decimals",
			rawPrice:      math.LegacyNewDec(1000),
			baseDecimals:  8,                         // BTC
			quoteDecimals: 6,                         // USDC
			expectedPrice: math.LegacyNewDec(100000), // 1000 * 10^(8-6)
		},
		{
			name:          "quote has more decimals",
			rawPrice:      math.LegacyNewDec(1000),
			baseDecimals:  6,                     // USDC
			quoteDecimals: 8,                     // BTC
			expectedPrice: math.LegacyNewDec(10), // 1000 / 10^(8-6)
		},
		{
			name:          "large base decimal",
			rawPrice:      math.LegacyNewDec(1),
			baseDecimals:  18,                               // ETH
			quoteDecimals: 6,                                // USDC
			expectedPrice: math.LegacyNewDec(1000000000000), // 1 * 10^(18-6)
		},
		{
			name:          "large quote decimal",
			rawPrice:      math.LegacyNewDec(1000000000000),
			baseDecimals:  6,                    // USDC
			quoteDecimals: 18,                   // ETH
			expectedPrice: math.LegacyNewDec(1), // 1000000000000 / 10^(18-6)
		},
		{
			name:          "zero base decimals",
			rawPrice:      math.LegacyNewDec(100),
			baseDecimals:  0,
			quoteDecimals: 6,
			expectedPrice: math.LegacyNewDec(1).Quo(math.LegacyNewDec(1000000)), // 100 / 10^6
		},
		{
			name:          "zero quote decimals",
			rawPrice:      math.LegacyNewDec(100),
			baseDecimals:  6,
			quoteDecimals: 0,
			expectedPrice: math.LegacyNewDec(100000000), // 100 * 10^6
		},
		{
			name:          "both zero decimals",
			rawPrice:      math.LegacyNewDec(100),
			baseDecimals:  0,
			quoteDecimals: 0,
			expectedPrice: math.LegacyNewDec(100),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			adjustedPrice := keeper.AdjustSpotPriceForDecimals(
				tc.rawPrice,
				tc.baseDecimals,
				tc.quoteDecimals,
			)

			s.Require().InDelta(
				tc.expectedPrice.MustFloat64(),
				adjustedPrice.MustFloat64(),
				0.0001,
				"expected price %v, got %v", tc.expectedPrice, adjustedPrice,
			)
		})
	}
}
