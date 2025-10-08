package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

func (s *KeeperTestSuite) CreateTradeRoutes() (routes []types.TradeRoute) {
	for i := 1; i <= 5; i++ {
		hostChain := fmt.Sprintf("chain-H%d", i)
		rewardChain := fmt.Sprintf("chain-R%d", i)
		tradeChain := fmt.Sprintf("chain-T%d", i)

		hostICA := types.ICAAccount{
			ChainId:      hostChain,
			Type:         types.ICAAccountType_WITHDRAWAL,
			ConnectionId: fmt.Sprintf("connection-0%d", i),
			Address:      "host_ica_address",
		}
		rewardICA := types.ICAAccount{
			ChainId:      rewardChain,
			Type:         types.ICAAccountType_CONVERTER_UNWIND,
			ConnectionId: fmt.Sprintf("connection-1%d", i),
			Address:      "reward_ica_address",
		}
		tradeICA := types.ICAAccount{
			ChainId:      tradeChain,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
			ConnectionId: fmt.Sprintf("connection-2%d", i),
			Address:      "trade_ica_address",
		}

		hostDenom := fmt.Sprintf("host-denom-%d", i)
		rewardDenom := fmt.Sprintf("reward-denom-%d", i)

		route := types.TradeRoute{
			RewardDenomOnHostZone:   "ibc-" + rewardDenom + "-on-" + hostChain,
			RewardDenomOnRewardZone: rewardDenom,
			RewardDenomOnTradeZone:  "ibc-" + rewardDenom + "-on-" + tradeChain,
			HostDenomOnTradeZone:    "ibc-" + hostDenom + "-on-" + tradeChain,
			HostDenomOnHostZone:     hostDenom,

			HostAccount:   hostICA,
			RewardAccount: rewardICA,
			TradeAccount:  tradeICA,

			HostToRewardChannelId:  fmt.Sprintf("channel-0%d", i),
			RewardToTradeChannelId: fmt.Sprintf("channel-1%d", i),
			TradeToHostChannelId:   fmt.Sprintf("channel-2%d", i),

			MinTransferAmount: sdkmath.ZeroInt(),

			// TradeConfig is deprecated but we include it so that we can compare with Equals
			// which would fail otherwise due to uninitialized types
			TradeConfig: types.TradeConfig{ //nolint:staticcheck
				SwapPrice:              sdkmath.LegacyZeroDec(),
				MaxAllowedSwapLossRate: sdkmath.LegacyZeroDec(),
				MinSwapAmount:          sdkmath.ZeroInt(),
				MaxSwapAmount:          sdkmath.ZeroInt(),
			},
		}
		routes = append(routes, route)

		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)
	}

	return routes
}

func (s *KeeperTestSuite) TestGetTradeRoute() {
	routes := s.CreateTradeRoutes()
	for i, route := range routes {
		rewardDenom := route.RewardDenomOnRewardZone
		hostDenom := route.HostDenomOnHostZone

		actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, rewardDenom, hostDenom)
		s.Require().True(found, "route should have been found")
		s.Require().Equal(routes[i], actualRoute, "route doesn't match")
	}
}

func (s *KeeperTestSuite) TestRemoveTradeRoute() {
	routes := s.CreateTradeRoutes()
	for _, route := range routes {
		s.App.StakeibcKeeper.RemoveTradeRoute(s.Ctx, route.RewardDenomOnRewardZone, route.HostDenomOnHostZone)
		_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, route.RewardDenomOnRewardZone, route.HostDenomOnHostZone)
		s.Require().False(found, "route should not have been found")
	}
}

func (s *KeeperTestSuite) TestGetAllTradeRoutes() {
	expectedRoutes := s.CreateTradeRoutes()
	actualRoutes := s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx)
	s.Require().ElementsMatch(expectedRoutes, actualRoutes)
}

func (s *KeeperTestSuite) TestGetTradeRouteFromTradeAccountChainId() {
	// Store 3 trade routes
	for i := 1; i <= 3; i++ {
		rewardDenom := fmt.Sprintf("reward-%d", i)
		hostDenom := fmt.Sprintf("host-%d", i)
		chainId := fmt.Sprintf("chain-%d", i)

		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, types.TradeRoute{
			RewardDenomOnRewardZone: rewardDenom,
			HostDenomOnHostZone:     hostDenom,
			TradeAccount: types.ICAAccount{
				ChainId: chainId,
			},
		})
	}

	// Search for each of them by chain ID
	for i := 1; i <= 3; i++ {
		rewardDenom := fmt.Sprintf("reward-%d", i)
		hostDenom := fmt.Sprintf("host-%d", i)
		chainId := fmt.Sprintf("chain-%d", i)

		actualRoute, found := s.App.StakeibcKeeper.GetTradeRouteFromTradeAccountChainId(s.Ctx, chainId)
		s.Require().True(found, "trade route 1 should have been found")
		s.Require().Equal(actualRoute.RewardDenomOnRewardZone, rewardDenom, "reward denom")
		s.Require().Equal(actualRoute.HostDenomOnHostZone, hostDenom, "host denom")
	}

	// Search for a chainId without a trade route
	_, found := s.App.StakeibcKeeper.GetTradeRouteFromTradeAccountChainId(s.Ctx, "chain-4")
	s.Require().False(found, "trade route should not have been found")
}
