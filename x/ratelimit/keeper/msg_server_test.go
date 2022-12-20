package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	minttypes "github.com/Stride-Labs/stride/v4/x/mint/types"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var (
	addRateLimitMsg = &types.MsgAddRateLimit{
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: 10,
		MaxPercentSend: 20,
		DurationHours:  30,
	}

	updateRateLimitMsg = &types.MsgUpdateRateLimit{
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: 20,
		MaxPercentSend: 30,
		DurationHours:  40,
	}

	removeRateLimitMsg = &types.MsgRemoveRateLimit{
		Denom:     "denom",
		ChannelId: "channel-0",
	}

	resetRateLimitMsg = &types.MsgResetRateLimit{
		Denom:     "denom",
		ChannelId: "channel-0",
	}
)

func (s *KeeperTestSuite) TestMsgServer_AddRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr

	denom := addRateLimitMsg.Denom
	channelId := addRateLimitMsg.ChannelId
	channelValue := int64(100)

	// Mint tokens for generating channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(addRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Check for duplicate rate limit
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitKeyAlreadyExists)
}

func (s *KeeperTestSuite) TestMsgServer_UpdateRateLimit() {
	s.SetupTest()
	validAddr, _ := apptesting.GenerateTestAddrs()
	addRateLimitMsg.Creator = validAddr
	updateRateLimitMsg.Creator = validAddr

	denom := updateRateLimitMsg.Denom
	channelId := updateRateLimitMsg.ChannelId
	channelValue := int64(100)

	// Mint tokens for generating channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(updateRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)

	// Attempt to update a rate limit that does not exist
	_, err = s.GetMsgServer().UpdateRateLimit(sdk.WrapSDKContext(s.Ctx), updateRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitKeyNotFound)

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
	channelValue := int64(100)

	// Mint tokens for generating channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(removeRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)

	// Attempt to remove a rate limit that does not exist
	_, err = s.GetMsgServer().RemoveRateLimit(sdk.WrapSDKContext(s.Ctx), removeRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitKeyNotFound)

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
	channelValue := int64(100)

	// Mint tokens for generating channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(resetRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)

	// Attempt to reset a rate limit that does not exist
	_, err = s.GetMsgServer().ResetRateLimit(sdk.WrapSDKContext(s.Ctx), resetRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitKeyNotFound)

	// Add a rate limit successfully
	_, err = s.GetMsgServer().AddRateLimit(sdk.WrapSDKContext(s.Ctx), addRateLimitMsg)
	s.Require().NoError(err)

	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)

	// Reset the rate limit successfully
	_, err = s.GetMsgServer().ResetRateLimit(sdk.WrapSDKContext(s.Ctx), resetRateLimitMsg)
	s.Require().NoError(err)

	// Check ratelimit quota is flow correctly
	resetRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)
	s.Require().Equal(resetRateLimit.Flow, &types.Flow{
		Inflow:       0,
		Outflow:      0,
		ChannelValue: uint64(channelValue),
	})
}
