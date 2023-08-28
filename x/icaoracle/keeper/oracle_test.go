package keeper_test

import (
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestGetOracle() {
	oracles := s.CreateTestOracles()

	expectedOracle := oracles[1]

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, expectedOracle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().Equal(expectedOracle, actualOracle)
}

func (s *KeeperTestSuite) TestGetAllOracles() {
	expectedOracles := s.CreateTestOracles()
	actualOracles := s.App.ICAOracleKeeper.GetAllOracles(s.Ctx)
	s.Require().Len(actualOracles, len(expectedOracles), "number of oracles")
	s.Require().ElementsMatch(expectedOracles, actualOracles, "contents of oracles")
}

func (s *KeeperTestSuite) TestRemoveOracle() {
	oracles := s.CreateTestOracles()

	oracleToRemove := oracles[1]

	// Remove the oracle
	s.App.ICAOracleKeeper.RemoveOracle(s.Ctx, oracleToRemove.ChainId)
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToRemove.ChainId)
	s.Require().False(found, "the removed oracle should not have been found, but it was")
}

func (s *KeeperTestSuite) TestToggleOracle() {
	oracles := s.CreateTestOracles()
	oracleToToggle := oracles[1]

	// Set the oracle to inactive
	err := s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, false)
	s.Require().NoError(err, "no error expected when toggling oracle")

	oracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToToggle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().False(oracle.Active, "oracle should have been marked inactive")

	// Remove the oracle connection ID and then try to re-activate it, it should fail
	invalidOracle := oracleToToggle
	invalidOracle.ConnectionId = ""
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, invalidOracle)

	err = s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, true)
	s.Require().ErrorContains(err, "oracle ICA channel has not been registered")

	// Remove the oracle contract address and try to re-activate it, it should fail
	invalidOracle = oracleToToggle
	invalidOracle.ContractAddress = ""
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, invalidOracle)

	err = s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, true)
	s.Require().ErrorContains(err, "oracle not instantiated")

	// Reset the oracle with all fields present
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracleToToggle)

	// Close the ICA channel and try to re-active it, it should fail
	s.UpdateChannelState(oracle.PortId, oracle.ChannelId, channeltypes.CLOSED)

	err = s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, true)
	s.Require().ErrorContains(err, "oracle ICA channel is closed")

	// Re-open the channel and try once more - this time it should succeed
	s.UpdateChannelState(oracle.PortId, oracle.ChannelId, channeltypes.OPEN)

	err = s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, true)
	s.Require().NoError(err, "no error expected when toggling oracle")

	oracle, found = s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToToggle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().True(oracle.Active, "oracle should have been marked as active")
}

func (s *KeeperTestSuite) TestGetOracleFromConnectionId() {
	oracles := s.CreateTestOracles()

	// Get oracle using connection Id
	expectedOracle := oracles[1]
	actualOracle, found := s.App.ICAOracleKeeper.GetOracleFromConnectionId(s.Ctx, expectedOracle.ConnectionId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().Equal(expectedOracle, actualOracle)

	// Attempt to get an oracle with a fake connectionId - should fail
	_, found = s.App.ICAOracleKeeper.GetOracleFromConnectionId(s.Ctx, "fake-connection-id")
	s.Require().False(found, "oracle should not have been found, but it was")
}

func (s *KeeperTestSuite) TestIsOracleICAChannelOpen() {
	s.CreateTransferChannel("chain-1")

	oracle := types.Oracle{
		PortId:    transfertypes.PortID,
		ChannelId: ibctesting.FirstChannelID,
	}

	// Check with an open channel, should equal true
	isOpen := s.App.ICAOracleKeeper.IsOracleICAChannelOpen(s.Ctx, oracle)
	s.Require().True(isOpen, "channel should be open")

	// Close the channel
	s.UpdateChannelState(transfertypes.PortID, ibctesting.FirstChannelID, channeltypes.CLOSED)

	// Try again, it should be false this time
	isOpen = s.App.ICAOracleKeeper.IsOracleICAChannelOpen(s.Ctx, oracle)
	s.Require().False(isOpen, "channel should now be closed")

	// Try with a fake channel
	oracle.ChannelId = "fake_channel"
	isOpen = s.App.ICAOracleKeeper.IsOracleICAChannelOpen(s.Ctx, oracle)
	s.Require().False(isOpen, "channel does not exist")
}
