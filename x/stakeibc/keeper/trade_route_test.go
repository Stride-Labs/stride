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
			ChainId: hostChain,
			Type:    types.ICAAccountType_WITHDRAWAL,
		}
		rewardICA := types.ICAAccount{
			ChainId: rewardChain,
			Type:    types.ICAAccountType_UNWIND,
		}
		tradeICA := types.ICAAccount{
			ChainId: tradeChain,
			Type:    types.ICAAccountType_TRADE,
		}

		hostRewardHop := types.TradeHop{
			FromAccount: hostICA,
			ToAccount:   rewardICA,
		}
		rewardTradeHop := types.TradeHop{
			FromAccount: rewardICA,
			ToAccount:   tradeICA,
		}
		tradeHostHop := types.TradeHop{
			FromAccount: tradeICA,
			ToAccount:   hostICA,
		}

		hostDenom := fmt.Sprintf("host-denom-%d", i)
		rewardDenom := fmt.Sprintf("reward-denom-%d", i)

		route := types.TradeRoute{
			RewardDenomOnHostZone:   "ibc-" + rewardDenom + "-on-" + hostChain,
			RewardDenomOnRewardZone: rewardDenom,
			RewardDenomOnTradeZone:  "ibc-" + rewardDenom + "-on-" + tradeChain,
			TargetDenomOnTradeZone:  "ibc-" + hostDenom + "-on-" + tradeChain,
			TargetDenomOnHostZone:   hostDenom,

			HostToRewardHop:  hostRewardHop,
			RewardToTradeHop: rewardTradeHop,
			TradeToHostHop:   tradeHostHop,

			PoolId:    uint64(i * 100),
			SpotPrice: "",

			MinSwapAmount: sdk.ZeroInt(),
			MaxSwapAmount: sdk.NewInt(1_000_000_000),
		}
		routes = append(routes, route)

		s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)
	}

	return routes
}

func (s *KeeperTestSuite) TestGetTradeRoute() {
	routes := s.CreateTradeRoutes()
	for i, route := range routes {
		startDenom := route.RewardDenomOnHostZone
		endDenom := route.TargetDenomOnHostZone

		actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, startDenom, endDenom)
		s.Require().True(found, "route should have been found")
		s.Require().Equal(routes[i], actualRoute, "route doesn't match")
	}
}

func (s *KeeperTestSuite) TestRemoveTradeRoute() {
	routes := s.CreateTradeRoutes()
	for _, route := range routes {
		s.App.StakeibcKeeper.RemoveTradeRoute(s.Ctx, route.RewardDenomOnHostZone, route.TargetDenomOnHostZone)
		_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, route.RewardDenomOnHostZone, route.TargetDenomOnHostZone)
		s.Require().False(found, "route should not have been found")
	}
}

func (s *KeeperTestSuite) TestGetAllTradeRoutes() {
	expectedRoutes := s.CreateTradeRoutes()
	actualRoutes := s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx)
	s.Require().ElementsMatch(expectedRoutes, actualRoutes)
}
