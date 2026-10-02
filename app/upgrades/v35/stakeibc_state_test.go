package v35_test

import (
	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestDeprecateComdex() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: v35.ComdexChainId, Halted: false, Deprecated: false})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: "juno-1", Deprecated: false})

	v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper)

	comdex, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().True(comdex.Deprecated, "comdex should be deprecated")
	s.Require().False(comdex.Halted, "halted is not touched")

	juno, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "juno-1")
	s.Require().False(juno.Deprecated, "other zones untouched")
}

func (s *UpgradeTestSuite) TestDeprecateComdex_MissingZone() {
	s.Require().NotPanics(func() { v35.DeprecateComdex(s.Ctx, s.App.StakeibcKeeper) })
}

func (s *UpgradeTestSuite) TestDeleteDydxTradeRoute() {
	dydx := stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom,
		HostDenomOnHostZone:     v35.DydxTradeRouteHostDenom,
	}
	other := stakeibctypes.TradeRoute{RewardDenomOnRewardZone: "uusdc", HostDenomOnHostZone: "uatom"}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, dydx)
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, other)

	v35.DeleteDydxTradeRoute(s.Ctx, s.App.StakeibcKeeper)

	_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, v35.DydxTradeRouteRewardDenom, v35.DydxTradeRouteHostDenom)
	s.Require().False(found, "dydx route deleted")
	s.Require().Len(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), 1, "the other route stays")
}

func (s *UpgradeTestSuite) TestDeleteDydxTradeRoute_MissingRoute() {
	s.Require().NotPanics(func() { v35.DeleteDydxTradeRoute(s.Ctx, s.App.StakeibcKeeper) })
}
