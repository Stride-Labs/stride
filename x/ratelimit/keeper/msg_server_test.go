package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var addRateLimitMsg = &types.MsgAddRateLimit{
	Denom:           "denom",
	ChannelId:       "channel-0",
	MaxPercentRecv:  10,
	MaxPercentSend:  20,
	DurationMinutes: 30,
}

var updateLimitMsg = &types.MsgUpdateRateLimit{
	PathId:          "denom/channel-0",
	MaxPercentRecv:  10,
	MaxPercentSend:  20,
	DurationMinutes: 30,
}

var removeRateLimitMsg = &types.MsgRemoveRateLimit{
	PathId: "denom/channel-0",
}

var resetRateLimitMsg = &types.MsgResetRateLimit{
	PathId: "denom/channel-0",
}

func (s *KeeperTestSuite) TestMsgServer_AddRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr

	// Add a rate limit successfully
	_, err := s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, "denom/channel-0")
	s.Require().True(found)

	// check for duplicate rate limit
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestMsgServer_RemoveRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()

	addRateLimitMsg.Creator = validAddr
	removeRateLimitMsg.Creator = validAddr
	pathId := removeRateLimitMsg.PathId

	// Attempt to remove a rate limit that does not exist
	_, err := s.GetMsgServer().RemoveRateLimit(sdk.WrapSDKContext(s.Ctx), removeRateLimitMsg)
	s.Require().Error(err)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, pathId)
	s.Require().True(found)

	// Remove the rate limit
	_, err = s.GetMsgServer().RemoveRateLimit(sdk.WrapSDKContext(s.Ctx), removeRateLimitMsg)
	s.Require().NoError(err)

	_, found = s.App.RatelimitKeeper.GetRateLimit(s.Ctx, pathId)
	s.Require().False(found)
}
