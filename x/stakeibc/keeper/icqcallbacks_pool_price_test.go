package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	icqtypes "github.com/Stride-Labs/stride/v21/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

type PoolPriceQueryCallbackTestCase struct {
	TradeRoute types.TradeRoute
	Price      sdk.Dec
	Response   ICQCallbackArgs
}

func (s *KeeperTestSuite) SetupPoolPriceCallbackTestCase(priceOnAsset0 bool) PoolPriceQueryCallbackTestCase {
	hostPrice := sdk.MustNewDecFromStr("1.2")
	rewardPrice := sdk.MustNewDecFromStr("0.8")

	// Alphabetize the and sort the denom's according to the ordering
	rewardDenom := "ibc/reward-denom-on-trade"
	hostDenom := "ibc/host-denom-on-trade"

	var twapRecord types.OsmosisTwapRecord
	if priceOnAsset0 {
		hostDenom = "a-" + hostDenom
		rewardDenom = "z-" + rewardDenom

		twapRecord = types.OsmosisTwapRecord{
			Asset0Denom:     hostDenom,
			Asset1Denom:     rewardDenom,
			P0LastSpotPrice: hostPrice,
			P1LastSpotPrice: rewardPrice,
		}
	} else {
		hostDenom = "z-" + hostDenom
		rewardDenom = "a-" + rewardDenom

		twapRecord = types.OsmosisTwapRecord{
			Asset0Denom:     rewardDenom,
			Asset1Denom:     hostDenom,
			P0LastSpotPrice: rewardPrice,
			P1LastSpotPrice: hostPrice,
		}
	}

	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,

		RewardDenomOnTradeZone: rewardDenom,
		HostDenomOnTradeZone:   hostDenom,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// Build query object and serialized query response
	callbackDataBz, _ := proto.Marshal(&types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	})
	query := icqtypes.Query{CallbackData: callbackDataBz}
	queryResponse, _ := proto.Marshal(&twapRecord)

	return PoolPriceQueryCallbackTestCase{
		TradeRoute: route,
		Price:      hostPrice,
		Response: ICQCallbackArgs{
			Query:        query,
			CallbackArgs: queryResponse,
		},
	}
}

func (s *KeeperTestSuite) TestPoolPriceCallback_Successful_HostTokenFirst() {
	hostTokenFirst := true
	tc := s.SetupPoolPriceCallbackTestCase(hostTokenFirst)

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().NoError(err)

	// Confirm the new price was set on the trade route
	route, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found")
	s.Require().Equal(tc.Price.String(), route.TradeConfig.SwapPrice.String(), "pool price")
}

func (s *KeeperTestSuite) TestPoolPriceCallback_Successful_RewardTokenFirst() {
	hostTokenFirst := false
	tc := s.SetupPoolPriceCallbackTestCase(hostTokenFirst)

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().NoError(err)

	// Confirm the new price was set on the trade route
	route, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found")
	s.Require().Equal(tc.Price.String(), route.TradeConfig.SwapPrice.String(), "pool price")
}

func (s *KeeperTestSuite) TestPoolPriceCallback_InvalidArgs() {
	tc := s.SetupPoolPriceCallbackTestCase(true) // ordering doesn't matter

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "unable to unmarshal the query response")
}

func (s *KeeperTestSuite) TestPoolPriceCallback_FailedToUnmarshalCallback() {
	tc := s.SetupPoolPriceCallbackTestCase(true) // ordering doesn't matter

	// Update the callback data so that it can't be successfully unmarshalled
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = []byte("random bytes")

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to unmarshal trade reward balance callback data")
}

func (s *KeeperTestSuite) TestPoolPriceCallback_TradeRouteNotFound() {
	tc := s.SetupPoolPriceCallbackTestCase(true) // ordering doesn't matter

	// Update the callback data so that it keys to a trade route that doesn't exist
	invalidCallbackDataBz, _ := proto.Marshal(&types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   "different-host-denom",
	})
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = invalidCallbackDataBz

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "trade route not found")
}

func (s *KeeperTestSuite) TestPoolPriceCallback_TradeDenomMismatch() {
	tc := s.SetupPoolPriceCallbackTestCase(true) // ordering doesn't matter

	// Update the trade route so that the denom's in the route don't match the query response
	invalidTradeRoute := tc.TradeRoute
	invalidTradeRoute.RewardDenomOnTradeZone = "different-denom"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, invalidTradeRoute)

	err := keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "Assets in query response")
	s.Require().ErrorContains(err, "do not match denom's from trade route")

	// Do it again, but with the other denom
	invalidTradeRoute = tc.TradeRoute
	invalidTradeRoute.HostDenomOnTradeZone = "different-denom"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, invalidTradeRoute)

	err = keeper.PoolPriceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "Assets in query response")
	s.Require().ErrorContains(err, "do not match denom's from trade route")
}

func (s *KeeperTestSuite) TestAssertTwapAssetsMatchTradeRoute() {
	testCases := []struct {
		name          string
		twapRecord    types.OsmosisTwapRecord
		tradeRoute    types.TradeRoute
		expectedMatch bool
	}{
		{
			name:          "successful match - 1",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-a", HostDenomOnTradeZone: "denom-b"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 2",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-b", Asset1Denom: "denom-a"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-b", HostDenomOnTradeZone: "denom-a"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 3",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-b", HostDenomOnTradeZone: "denom-a"},
			expectedMatch: true,
		},
		{
			name:          "successful match - 4",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-b", Asset1Denom: "denom-a"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-a", HostDenomOnTradeZone: "denom-b"},
			expectedMatch: true,
		},
		{
			name:          "mismatch osmo asset 0",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-z", Asset1Denom: "denom-b"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-a", HostDenomOnTradeZone: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch osmo asset 1",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-z"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-a", HostDenomOnTradeZone: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch route reward denom",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-z", HostDenomOnTradeZone: "denom-b"},
			expectedMatch: false,
		},
		{
			name:          "mismatch route host denom",
			twapRecord:    types.OsmosisTwapRecord{Asset0Denom: "denom-a", Asset1Denom: "denom-b"},
			tradeRoute:    types.TradeRoute{RewardDenomOnTradeZone: "denom-a", HostDenomOnTradeZone: "denom-z"},
			expectedMatch: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := keeper.AssertTwapAssetsMatchTradeRoute(tc.twapRecord, tc.tradeRoute)
			if tc.expectedMatch {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
