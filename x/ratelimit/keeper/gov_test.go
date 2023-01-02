package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	minttypes "github.com/Stride-Labs/stride/v4/x/mint/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var (
	addRateLimitMsg = types.AddRateLimitProposal{
		Title:          "AddRateLimit",
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: sdk.NewInt(10),
		MaxPercentSend: sdk.NewInt(20),
		DurationHours:  30,
	}

	updateRateLimitMsg = types.UpdateRateLimitProposal{
		Title:          "UpdateRateLimit",
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: sdk.NewInt(20),
		MaxPercentSend: sdk.NewInt(30),
		DurationHours:  40,
	}

	removeRateLimitMsg = types.RemoveRateLimitProposal{
		Title:     "RemoveRateLimit",
		Denom:     "denom",
		ChannelId: "channel-0",
	}

	resetRateLimitMsg = types.ResetRateLimitProposal{
		Title:     "ResetRateLimit",
		Denom:     "denom",
		ChannelId: "channel-0",
	}
)

// Helper function to create a channel and prevent a channel not exists error
func (s *KeeperTestSuite) createChannel(channelId string) {
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelId, channeltypes.Channel{})
}

// Helper function to mint tokens and create channel value to prevent a zero channel value error
func (s *KeeperTestSuite) createChannelValue(denom string, channelValue sdk.Int) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(addRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)
}

// Helper function to add a rate limit with an optional error expectation
func (s *KeeperTestSuite) addRateLimit(expectedErr *sdkerrors.Error) {
	actualErr := s.App.RatelimitKeeper.GovAddRateLimit(s.Ctx, &addRateLimitMsg)

	// If it should have been added successfully, confirm no error
	// and confirm the rate limit was created
	if expectedErr == nil {
		s.Require().NoError(actualErr)

		_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found)
	} else {
		// If it should have failed, check the error
		s.Require().Equal(actualErr, expectedErr)
	}
}

// Helper function to add a rate limit successfully
func (s *KeeperTestSuite) addRateLimitSuccessful() {
	s.addRateLimit(nil)
}

// Helper function to add a rate limit with an expected error
func (s *KeeperTestSuite) addRateLimitWithError(expectedErr *sdkerrors.Error) {
	s.addRateLimit(expectedErr)
}

func (s *KeeperTestSuite) TestMsgServer_AddRateLimit() {
	denom := addRateLimitMsg.Denom
	channelId := addRateLimitMsg.ChannelId
	channelValue := sdk.NewInt(100)

	// First try to add a rate limit when there's no channel value, it will fail
	s.addRateLimitWithError(types.ErrZeroChannelValue)

	// Create channel value
	s.createChannelValue(denom, channelValue)

	// Then try to add a rate limit before the channel has been created, it will also fail
	s.addRateLimitWithError(types.ErrChannelNotFound)

	// Create the channel
	s.createChannel(channelId)

	// Now add a rate limit successfully
	s.addRateLimitSuccessful()

	// Finally, try to add the same rate limit again - it should fail
	s.addRateLimitWithError(types.ErrRateLimitAlreadyExists)
}

func (s *KeeperTestSuite) TestMsgServer_UpdateRateLimit() {
	denom := updateRateLimitMsg.Denom
	channelId := updateRateLimitMsg.ChannelId
	channelValue := sdk.NewInt(100)

	// Create channel and channel value
	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to update a rate limit that does not exist
	err := s.App.RatelimitKeeper.GovUpdateRateLimit(s.Ctx, &updateRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Update the rate limit successfully
	err = s.App.RatelimitKeeper.GovUpdateRateLimit(s.Ctx, &updateRateLimitMsg)
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
	denom := removeRateLimitMsg.Denom
	channelId := removeRateLimitMsg.ChannelId
	channelValue := sdk.NewInt(100)

	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to remove a rate limit that does not exist
	err := s.App.RatelimitKeeper.GovRemoveRateLimit(s.Ctx, &removeRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Remove the rate limit successfully
	err = s.App.RatelimitKeeper.GovRemoveRateLimit(s.Ctx, &removeRateLimitMsg)
	s.Require().NoError(err)

	// Confirm it was removed
	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestMsgServer_ResetRateLimit() {
	denom := resetRateLimitMsg.Denom
	channelId := resetRateLimitMsg.ChannelId
	channelValue := sdk.NewInt(100)

	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to reset a rate limit that does not exist
	err := s.App.RatelimitKeeper.GovResetRateLimit(s.Ctx, &resetRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Reset the rate limit successfully
	err = s.App.RatelimitKeeper.GovResetRateLimit(s.Ctx, &resetRateLimitMsg)
	s.Require().NoError(err)

	// Check ratelimit quota is flow correctly
	resetRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)
	s.Require().Equal(resetRateLimit.Flow, &types.Flow{
		Inflow:       sdk.ZeroInt(),
		Outflow:      sdk.ZeroInt(),
		ChannelValue: channelValue,
	})
}
