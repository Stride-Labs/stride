package keeper_test

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/deps/osmomath"
	deps "github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types"
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

func (s *KeeperTestSuite) TestSubmitOsmosisClPoolICQUnknownPrice() {
	// Set up test parameters
	tokenPrice := types.TokenPrice{
		BaseDenom:     "uatom",
		QuoteDenom:    "uusdc",
		OsmosisPoolId: "1",
	}

	params := types.Params{
		OsmosisChainId:      "osmosis-1",
		OsmosisConnectionId: "connection-0",
		UpdateIntervalSec:   60,
	}
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Submit ICQ request
	err = s.App.ICQOracleKeeper.SubmitOsmosisClPoolICQ(s.Ctx, tokenPrice)
	s.Require().ErrorContains(err, "price not found")
}

func (s *KeeperTestSuite) TestHappyPathOsmosisClPoolICQ() {
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
		OsmosisPoolId: "1",
	}

	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	params := types.Params{
		OsmosisChainId:      "osmosis-1",
		OsmosisConnectionId: "connection-0",
		UpdateIntervalSec:   60,
	}
	err = s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Verify tokenPrice.QueryInProgress before
	s.Require().False(tokenPrice.QueryInProgress)

	// Submit ICQ request
	err = s.App.ICQOracleKeeper.SubmitOsmosisClPoolICQ(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Verify the submitted query
	s.Require().Equal(params.OsmosisChainId, submittedQuery.ChainId)
	s.Require().Equal(params.OsmosisConnectionId, submittedQuery.ConnectionId)
	s.Require().Equal(icqtypes.CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF, submittedQuery.QueryType)

	// Verify tokenPrice.QueryInProgress after
	tokenPriceAfter, err := s.App.ICQOracleKeeper.GetTokenPrice(s.Ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	s.Require().NoError(err)

	s.Require().True(tokenPriceAfter.QueryInProgress)
}

func (s *KeeperTestSuite) TestSubmitOsmosisClPoolICQBranches() {
	testCases := []struct {
		name          string
		setup         func()
		tokenPrice    types.TokenPrice
		expectedError string
	}{
		{
			name: "error parsing pool ID",
			setup: func() {
				// Set valid params
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "invalid_pool_id", // NaN invalid pool ID
			},
			expectedError: "Error converting osmosis pool id",
		},
		{
			name: "error submitting ICQ request",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

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
				OsmosisPoolId: "1",
			},
			expectedError: "Error submitting OsmosisClPool ICQ",
		},
		{
			name: "error setting query in progress",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

				// Setup mock ICQ keeper with success response
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
						// Remove token price so set query in progress will fail to get it after SubmitICQRequest returns
						s.App.ICQOracleKeeper.RemoveTokenPrice(ctx, "uatom", "uusdc", "1")
						return nil
					},
				}
				s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper

				// Don't set the token price first, which will cause SetTokenPriceQueryInProgress to fail
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "1",
			},
			expectedError: "Error updating queryInProgress=true",
		},
		{
			name: "successful submission",
			setup: func() {
				params := types.Params{
					OsmosisChainId:      "osmosis-1",
					OsmosisConnectionId: "connection-0",
					UpdateIntervalSec:   60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

				// Setup mock ICQ keeper with success response
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
						return nil
					},
				}
				s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper
			},
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "1",
			},
			expectedError: "",
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tc.tokenPrice)
				s.Require().NoError(err)
			}

			// Execute
			err := s.App.ICQOracleKeeper.SubmitOsmosisClPoolICQ(s.Ctx, tc.tokenPrice)

			// Verify results
			if tc.expectedError != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedError)
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

func (s *KeeperTestSuite) TestSubmitOsmosisClPoolICQQueryData() {
	var capturedQuery icqtypes.Query

	// Setup mock ICQ keeper to capture the submitted query and prove flag
	s.mockICQKeeper = MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			capturedQuery = query
			return nil
		},
	}
	s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper

	// Set up test parameters
	tokenPrice := types.TokenPrice{
		BaseDenom:     "uatom",
		QuoteDenom:    "uusdc",
		OsmosisPoolId: "1",
	}
	err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	params := types.Params{
		OsmosisChainId:      "osmosis-1",
		OsmosisConnectionId: "connection-0",
		UpdateIntervalSec:   60,
	}
	err = s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Submit ICQ request
	err = s.App.ICQOracleKeeper.SubmitOsmosisClPoolICQ(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Verify the captured query data
	s.Require().Equal(params.OsmosisChainId, capturedQuery.ChainId)
	s.Require().Equal(params.OsmosisConnectionId, capturedQuery.ConnectionId)
	s.Require().Equal(icqtypes.CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF, capturedQuery.QueryType)
	s.Require().Equal(types.ModuleName, capturedQuery.CallbackModule)
	s.Require().Equal(keeper.ICQCallbackID_OsmosisClPool, capturedQuery.CallbackId)

	// Verify request data format (pool key)
	osmosisPoolId, err := strconv.ParseUint(tokenPrice.OsmosisPoolId, 10, 64)
	s.Require().NoError(err)
	expectedRequestData := icqtypes.FormatOsmosisKeyPool(osmosisPoolId)
	s.Require().Equal(expectedRequestData, capturedQuery.RequestData)

	// Verify callback data contains the token price
	var decodedTokenPrice types.TokenPrice
	err = s.App.AppCodec().Unmarshal(capturedQuery.CallbackData, &decodedTokenPrice)
	s.Require().NoError(err)
	s.Require().Equal(tokenPrice.BaseDenom, decodedTokenPrice.BaseDenom)
	s.Require().Equal(tokenPrice.QuoteDenom, decodedTokenPrice.QuoteDenom)
	s.Require().Equal(tokenPrice.OsmosisPoolId, decodedTokenPrice.OsmosisPoolId)

	// Verify timeout settings
	expectedTimeout := time.Duration(params.UpdateIntervalSec) * time.Second
	s.Require().Equal(expectedTimeout, capturedQuery.TimeoutDuration)
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, capturedQuery.TimeoutPolicy)
}

// Helper function to create mock pool data
func (s *KeeperTestSuite) createMockPoolData(baseDenom string, quoteDenom string) []byte {
	pool := deps.OsmosisConcentratedLiquidityPool{
		Token0: baseDenom,
		Token1: quoteDenom,
		// The square root of the price
		// For a price of 1.5, 1.224744871 is the square root
		CurrentSqrtPrice: osmomath.NewBigDecWithPrec(1224744871, 9),
	}

	bz, err := proto.Marshal(&pool)
	s.Require().NoError(err, "no error expected when marshaling mock pool data")
	return bz
}

func (s *KeeperTestSuite) TestOsmosisClPoolCallback() {
	// Setup base test parameters used across test cases
	baseTokenPrice := types.TokenPrice{
		BaseDenom:         "uatom",
		QuoteDenom:        "uusdc",
		OsmosisPoolId:     "1",
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
				s.Require().NoError(err)

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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create mock pool data with expected price of 2.0
				poolData := s.createMockPoolData(
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
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.LastQueryTime.UnixNano())
				s.Require().InDelta(1.5, tokenPrice.SpotPrice.MustFloat64(), 0.01)
			},
		},
		{
			name: "successful update with inverse pool data",
			setup: func() ([]byte, icqtypes.Query) {
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with swapped denoms
				poolData := s.createMockPoolData(
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
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.LastQueryTime.UnixNano())
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
			expectedError: "price not found for baseDenom='' quoteDenom='' poolId=''",
		},
		{
			name: "nil query callback data",
			setup: func() ([]byte, icqtypes.Query) {
				return []byte{}, icqtypes.Query{
					CallbackData: nil,
				}
			},
			expectedError: "price not found for baseDenom='' quoteDenom='' poolId=''",
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with empty token denoms
				pool := deps.OsmosisConcentratedLiquidityPool{
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with invalid sqrt price
				pool := deps.OsmosisConcentratedLiquidityPool{
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create pool with invalid sqrt price
				pool := deps.OsmosisConcentratedLiquidityPool{
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
				err := s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, baseTokenPrice)
				s.Require().NoError(err)

				tokenPriceBz, err := s.App.AppCodec().Marshal(&baseTokenPrice)
				s.Require().NoError(err)

				// Create valid pool data
				poolData := s.createMockPoolData(
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
				"price not found for baseDenom='%s' quoteDenom='%s' poolId='%s'",
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
			err := s.icqCallbacks.CallICQCallback(
				s.Ctx,
				keeper.ICQCallbackID_OsmosisClPool,
				poolData,
				query,
			)

			// Verify results
			if tc.expectedError != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedError)
			} else if tc.verify != nil {
				tc.verify(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnmarshalSpotPriceFromOsmosisClPool() {
	baseTokenPrice := types.TokenPrice{
		BaseDenom:          "uatom",
		QuoteDenom:         "uusdc",
		OsmosisPoolId:      "1",
		OsmosisBaseDenom:   "ibc/uatom",
		OsmosisQuoteDenom:  "uusdc",
		BaseDenomDecimals:  6,
		QuoteDenomDecimals: 6,
	}

	testCases := []struct {
		name          string
		tokenPrice    types.TokenPrice
		poolData      []byte
		expectedPrice math.LegacyDec
		expectedError string
	}{
		{
			name:          "invalid pool data",
			tokenPrice:    baseTokenPrice,
			poolData:      []byte("invalid pool data"),
			expectedError: "proto: wrong wireType",
		},
		{
			name:       "successful price calculation with equal decimals",
			tokenPrice: baseTokenPrice,
			poolData: s.createMockPoolData(
				baseTokenPrice.OsmosisBaseDenom,
				baseTokenPrice.OsmosisQuoteDenom,
			),
			expectedPrice: math.LegacyNewDecWithPrec(15, 1), // 1.5 from mock pool data
		},
		{
			name: "successful price calculation with more base decimals",
			tokenPrice: types.TokenPrice{
				BaseDenom:          "satoshi",
				QuoteDenom:         "uusdc",
				OsmosisPoolId:      "1",
				OsmosisBaseDenom:   "ibc/satoshi",
				OsmosisQuoteDenom:  "ibc/uusdc",
				BaseDenomDecimals:  8, // BTC has 8 decimals
				QuoteDenomDecimals: 6, // USDC has 6 decimals
			},
			poolData: s.createMockPoolData(
				"ibc/satoshi",
				"ibc/uusdc",
			),
			expectedPrice: math.LegacyNewDecWithPrec(15, 1).Mul(math.LegacyNewDec(100)), // 1.5 * 10^2
		},
		{
			name: "successful price calculation with more quote decimals",
			tokenPrice: types.TokenPrice{
				BaseDenom:          "uatom",
				QuoteDenom:         "uusdc",
				OsmosisPoolId:      "1",
				OsmosisBaseDenom:   "ibc/uatom",
				OsmosisQuoteDenom:  "ibc/uusdc",
				BaseDenomDecimals:  6, // ATOM has 6 decimals
				QuoteDenomDecimals: 8, // Quote has 8 decimals
			},
			poolData: s.createMockPoolData(
				"ibc/uatom",
				"ibc/uusdc",
			),
			expectedPrice: math.LegacyNewDecWithPrec(15, 1).Quo(math.LegacyNewDec(100)), // 1.5 / 10^2
		},
		{
			name: "price calculation with inverse pool ordering",
			tokenPrice: types.TokenPrice{
				BaseDenom:          "uatom",
				QuoteDenom:         "uusdc",
				OsmosisPoolId:      "1",
				OsmosisBaseDenom:   "ibc/uatom",
				OsmosisQuoteDenom:  "ibc/uusdc",
				BaseDenomDecimals:  8,
				QuoteDenomDecimals: 6,
			},
			poolData: s.createMockPoolData(
				"ibc/uusdc",
				"ibc/uatom",
			),
			expectedPrice: math.LegacyMustNewDecFromStr("0.666666666666666667").Mul(math.LegacyNewDec(100)), // (1/1.5) * 10^2
		},
		{
			name: "different denom ordering in pool",
			tokenPrice: types.TokenPrice{
				BaseDenom:          "uatom",
				QuoteDenom:         "uusdc",
				OsmosisPoolId:      "1",
				OsmosisBaseDenom:   "ibc/uatom",
				OsmosisQuoteDenom:  "different_denom",
				BaseDenomDecimals:  6,
				QuoteDenomDecimals: 6,
			},
			poolData: s.createMockPoolData(
				"ibc/uatom",
				"uusdc",
			),
			expectedError: "quote asset denom (different_denom) is not in pool with (ibc/uatom, uusdc) pair",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			spotPrice, err := keeper.UnmarshalSpotPriceFromOsmosisClPool(tc.tokenPrice, tc.poolData)

			if tc.expectedError != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectedError)
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
