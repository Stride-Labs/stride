package keeper_test

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/deps/osmomath"
	deps "github.com/Stride-Labs/stride/v24/x/icqoracle/deps/types"
	"github.com/Stride-Labs/stride/v24/x/icqoracle/keeper"
	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v24/x/interchainquery/types"
)

// Mock ICQ Keeper struct
type MockICQKeeper struct {
	SubmitICQRequestFn func(ctx sdk.Context, query icqtypes.Query, prove bool) error
}

func (m MockICQKeeper) SubmitICQRequest(ctx sdk.Context, query icqtypes.Query, prove bool) error {
	if m.SubmitICQRequestFn != nil {
		return m.SubmitICQRequestFn(ctx, query, prove)
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
		IcqTimeoutSec:       60,
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
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, prove bool) error {
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
		IcqTimeoutSec:       60,
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
					IcqTimeoutSec:       60,
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
					IcqTimeoutSec:       60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

				// Mock ICQ keeper to return error
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, prove bool) error {
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
					IcqTimeoutSec:       60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

				// Setup mock ICQ keeper with success response
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, prove bool) error {
						// Remove token price so set query in progress will fail to get it after SubmitICQRequestFn returns
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
					IcqTimeoutSec:       60,
				}
				err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
				s.Require().NoError(err)

				// Setup mock ICQ keeper with success response
				s.mockICQKeeper = MockICQKeeper{
					SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, prove bool) error {
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
	var capturedProve bool

	// Setup mock ICQ keeper to capture the submitted query and prove flag
	s.mockICQKeeper = MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, prove bool) error {
			capturedQuery = query
			capturedProve = prove
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
		IcqTimeoutSec:       60,
	}
	err = s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Submit ICQ request
	err = s.App.ICQOracleKeeper.SubmitOsmosisClPoolICQ(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Verify the captured query data
	s.Require().NotEmpty(capturedQuery.Id, "query ID should not be empty")
	expectedQueryId := fmt.Sprintf("%s|%s|%s|%d",
		tokenPrice.BaseDenom,
		tokenPrice.QuoteDenom,
		tokenPrice.OsmosisPoolId,
		s.Ctx.BlockHeight(),
	)
	s.Require().Equal(expectedQueryId, capturedQuery.Id, "query ID should match expected format")

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
	expectedTimeout := time.Duration(params.IcqTimeoutSec) * time.Second
	s.Require().Equal(expectedTimeout, capturedQuery.TimeoutDuration)
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, capturedQuery.TimeoutPolicy)

	// Verify prove flag
	// For Osmosis CL pool queries, we require cryptographic proofs to ensure data authenticity
	s.Require().True(capturedProve, "prove flag should be true to request cryptographic proofs in ICQ response")
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
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.UpdatedAt.UnixNano())
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
				s.Require().Equal(s.Ctx.BlockTime().UnixNano(), tokenPrice.UpdatedAt.UnixNano())
				s.Require().InDelta(1/1.5, tokenPrice.SpotPrice.MustFloat64(), 0.01) // inversed price
			},
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
		BaseDenom:         "uatom",
		QuoteDenom:        "uusdc",
		OsmosisPoolId:     "1",
		OsmosisBaseDenom:  "ibc/uatom",
		OsmosisQuoteDenom: "uusdc",
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
			expectedError: "failed to unmarshal",
		},
		{
			name:       "successful price calculation",
			tokenPrice: baseTokenPrice,
			poolData: s.createMockPoolData(
				baseTokenPrice.OsmosisBaseDenom,
				baseTokenPrice.OsmosisQuoteDenom,
			),
			expectedPrice: math.LegacyNewDec(1), // Based on our mock pool data
		},
		{
			name:       "swapped denoms",
			tokenPrice: baseTokenPrice,
			poolData: s.createMockPoolData(
				baseTokenPrice.OsmosisQuoteDenom, // Swapped!
				baseTokenPrice.OsmosisBaseDenom,  // Swapped!
			),
			expectedError: "denom not found in pool",
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
				s.Require().True(tc.expectedPrice.Sub(spotPrice).Abs().LTE(math.LegacyNewDecWithPrec(1, 6)))
			}
		})
	}
}
