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

// Helper function to create mock pool data
func (s *KeeperTestSuite) createMockPoolData(baseDenom string, quoteDenom string) []byte {
	pool := deps.OsmosisConcentratedLiquidityPool{
		Id:                   1,
		CurrentTickLiquidity: math.LegacyNewDec(1000000),
		Token0:               baseDenom,
		Token1:               quoteDenom,
		CurrentSqrtPrice:     osmomath.NewBigDec(1000000), // This represents a price of 1.0
		CurrentTick:          0,
		TickSpacing:          100,
		ExponentAtPriceOne:   6,
		SpreadFactor:         math.LegacyNewDecWithPrec(1, 3), // 0.1%
	}

	bz, err := proto.Marshal(&pool)
	s.Require().NoError(err, "no error expected when marshaling mock pool data")
	return bz
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

func (s *KeeperTestSuite) TestICQCallbacksRegistration() {
	callbacks := s.App.ICQOracleKeeper.ICQCallbackHandler().RegisterICQCallbacks()

	// Verify the expected callbacks are registered
	s.Require().True(callbacks.HasICQCallback("osmosisclpool"), "osmosisclpool callback should be registered after RegisterICQCallbacks")
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
