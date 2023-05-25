package gov_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	minttypes "github.com/Stride-Labs/stride/v9/x/mint/types"
	"github.com/Stride-Labs/stride/v9/x/ratelimit/keeper/gov"
	"github.com/Stride-Labs/stride/v9/x/ratelimit/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

var (
	addRateLimitMsg = types.AddRateLimitProposal{
		Title:          "AddRateLimit",
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: sdkmath.NewInt(10),
		MaxPercentSend: sdkmath.NewInt(20),
		DurationHours:  30,
	}

	updateRateLimitMsg = types.UpdateRateLimitProposal{
		Title:          "UpdateRateLimit",
		Denom:          "denom",
		ChannelId:      "channel-0",
		MaxPercentRecv: sdkmath.NewInt(20),
		MaxPercentSend: sdkmath.NewInt(30),
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
func (s *KeeperTestSuite) createChannelValue(denom string, channelValue sdkmath.Int) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(addRateLimitMsg.Denom, channelValue)))
	s.Require().NoError(err)
}

// Helper function to add a rate limit with an optional error expectation
func (s *KeeperTestSuite) addRateLimit(expectedErr *errorsmod.Error) {
	actualErr := gov.AddRateLimit(s.Ctx, s.App.RatelimitKeeper, s.App.IBCKeeper.ChannelKeeper, &addRateLimitMsg)

	// If it should have been added successfully, confirm no error
	// and confirm the rate limit was created
	if expectedErr == nil {
		s.Require().NoError(actualErr)

		_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, addRateLimitMsg.Denom, addRateLimitMsg.ChannelId)
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
func (s *KeeperTestSuite) addRateLimitWithError(expectedErr *errorsmod.Error) {
	s.addRateLimit(expectedErr)
}

func (s *KeeperTestSuite) TestMsgServer_AddRateLimit() {
	denom := addRateLimitMsg.Denom
	channelId := addRateLimitMsg.ChannelId
	channelValue := sdkmath.NewInt(100)

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
	channelValue := sdkmath.NewInt(100)

	// Create channel and channel value
	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to update a rate limit that does not exist
	err := gov.UpdateRateLimit(s.Ctx, s.App.RatelimitKeeper, &updateRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Update the rate limit successfully
	err = gov.UpdateRateLimit(s.Ctx, s.App.RatelimitKeeper, &updateRateLimitMsg)
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
	channelValue := sdkmath.NewInt(100)

	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to remove a rate limit that does not exist
	err := gov.RemoveRateLimit(s.Ctx, s.App.RatelimitKeeper, &removeRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Remove the rate limit successfully
	err = gov.RemoveRateLimit(s.Ctx, s.App.RatelimitKeeper, &removeRateLimitMsg)
	s.Require().NoError(err)

	// Confirm it was removed
	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestMsgServer_ResetRateLimit() {
	denom := resetRateLimitMsg.Denom
	channelId := resetRateLimitMsg.ChannelId
	channelValue := sdkmath.NewInt(100)

	s.createChannel(channelId)
	s.createChannelValue(denom, channelValue)

	// Attempt to reset a rate limit that does not exist
	err := gov.ResetRateLimit(s.Ctx, s.App.RatelimitKeeper, &resetRateLimitMsg)
	s.Require().Equal(err, types.ErrRateLimitNotFound)

	// Add a rate limit successfully
	s.addRateLimitSuccessful()

	// Reset the rate limit successfully
	err = gov.ResetRateLimit(s.Ctx, s.App.RatelimitKeeper, &resetRateLimitMsg)
	s.Require().NoError(err)

	// Check ratelimit quota is flow correctly
	resetRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found)
	s.Require().Equal(resetRateLimit.Flow, &types.Flow{
		Inflow:       sdkmath.ZeroInt(),
		Outflow:      sdkmath.ZeroInt(),
		ChannelValue: channelValue,
	})
}
