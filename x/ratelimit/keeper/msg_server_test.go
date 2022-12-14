package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var addRateLimitMsg = &types.MsgAddRateLimit{
	Denom:          "denom",
	ChannelId:      "channel-0",
	MaxPercentRecv: 10,
	MaxPercentSend: 20,
	DurationHours:  30,
}

var updateRateLimitMsg = &types.MsgUpdateRateLimit{
	Denom:          "denom",
	ChannelId:      "channel-0",
	MaxPercentRecv: 20,
	MaxPercentSend: 30,
	DurationHours:  40,
}

var removeRateLimitMsg = &types.MsgRemoveRateLimit{
	Denom:     "denom",
	ChannelId: "channel-0",
}

var resetRateLimitMsg = &types.MsgResetRateLimit{
	Denom:     "denom",
	ChannelId: "channel-0",
}

func (s *KeeperTestSuite) TestMsgServer_AddRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr

	denom := addRateLimitMsg.Denom
	channelId := addRateLimitMsg.ChannelId

	// Add a rate limit successfully
	_, err := s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Check for duplicate rate limit
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestMsgServer_UpdateRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr
	updateRateLimitMsg.Creator = validAddr

	denom := updateRateLimitMsg.Denom
	channelId := updateRateLimitMsg.ChannelId

	// Attempt to update a rate limit that does not exist
	_, err := s.GetMsgServer().UpdateRateLimit(sdk.WrapSDKContext(s.Ctx), updateRateLimitMsg)
	s.Require().Error(err)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Update the rate limit successfully
	_, err = s.GetMsgServer().UpdateRateLimit(sdk.WrapSDKContext(s.Ctx), updateRateLimitMsg)
	s.Require().NoError(err)

	// Check ratelimit quota is updated correctly
	updatedRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)
	s.Require().Equal(updatedRateLimit.Quota, &types.Quota{
		MaxPercentSend: updateRateLimitMsg.MaxPercentSend,
		MaxPercentRecv: updateRateLimitMsg.MaxPercentRecv,
		DurationHours:  updateRateLimitMsg.DurationHours,
	})
}

func (s *KeeperTestSuite) TestMsgServer_RemoveRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()

	addRateLimitMsg.Creator = validAddr
	removeRateLimitMsg.Creator = validAddr
	denom := removeRateLimitMsg.Denom
	channelId := removeRateLimitMsg.ChannelId

	// Attempt to remove a rate limit that does not exist
	_, err := s.GetMsgServer().RemoveRateLimit(sdk.WrapSDKContext(s.Ctx), removeRateLimitMsg)
	s.Require().Error(err)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Remove the rate limit
	_, err = s.GetMsgServer().RemoveRateLimit(sdk.WrapSDKContext(s.Ctx), removeRateLimitMsg)
	s.Require().NoError(err)

	_, found = s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestMsgServer_ResetRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr
	resetRateLimitMsg.Creator = validAddr

	denom := resetRateLimitMsg.Denom
	channelId := resetRateLimitMsg.ChannelId

	// Attempt to reset a rate limit that does not exist
	_, err := s.GetMsgServer().ResetRateLimit(sdk.WrapSDKContext(s.Ctx), resetRateLimitMsg)
	s.Require().Error(err)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Reset the rate limit successfully
	_, err = s.GetMsgServer().ResetRateLimit(sdk.WrapSDKContext(s.Ctx), resetRateLimitMsg)
	s.Require().NoError(err)

	// Check ratelimit quota is reset correctly
	resetRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)
	s.Require().Equal(resetRateLimit.Flow.Inflow, uint64(0))
	s.Require().Equal(resetRateLimit.Flow.Outflow, uint64(0))
}
