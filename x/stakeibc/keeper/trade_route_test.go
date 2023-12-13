package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
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

		tradeConfig := types.TradeConfig{
			PoolId:                 uint64(i * 100),
			SwapPrice:              sdk.OneDec(),
			MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.05"),

			MinSwapAmount: sdk.ZeroInt(),
			MaxSwapAmount: sdk.NewInt(1_000_000_000),
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

			TradeConfig: tradeConfig,
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
