package keeper_test

import (
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestDeleteTradeRoute() {
	initialRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, initialRoute)

	msg := types.MsgDeleteTradeRoute{
		Authority:   Authority,
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	}

	// Confirm the route is present before attepmting to delete was deleted
	_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found before delete message")

	// Delete the trade route
	_, err := s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when deleting trade route")

	// Confirm it was deleted
	_, found = s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().False(found, "trade route should have been deleted")

	// Attempt to delete it again, it should fail since it doesn't exist
	_, err = s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "trade route not found")

	// Attempt to delete with the wrong authority - it should fail
	invalidMsg := msg
	invalidMsg.Authority = "not-gov-address"

	_, err = s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid authority")
}
