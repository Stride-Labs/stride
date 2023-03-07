package keeper_test

import (
	"strconv"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Helper function to create 5 oracle objects with various attributes
func (s *KeeperTestSuite) createOracles() []types.Oracle {
	oracles := []types.Oracle{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		oracle := types.Oracle{
			ChainId:      "chain-" + suffix,
			ConnectionId: "connection-" + suffix,
			Active:       true,
		}

		oracles = append(oracles, oracle)
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}
	return oracles
}

func (s *KeeperTestSuite) TestGetOracle() {
	oracles := s.createOracles()

	expectedOracle := oracles[1]

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, expectedOracle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().Equal(expectedOracle, actualOracle)
}

func (s *KeeperTestSuite) TestGetAllOracles() {
	expectedOracles := s.createOracles()
	actualOracles := s.App.ICAOracleKeeper.GetAllOracles(s.Ctx)
	s.Require().Len(actualOracles, len(expectedOracles), "number of oracles")
	s.Require().ElementsMatch(expectedOracles, actualOracles, "contents of oracles")
}

func (s *KeeperTestSuite) TestRemoveOracle() {
	oracles := s.createOracles()

	oracleToRemove := oracles[1]

	// Remove the oracle
	s.App.ICAOracleKeeper.RemoveOracle(s.Ctx, oracleToRemove.ChainId)
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToRemove.ChainId)
	s.Require().False(found, "the removed oracle should not have been found, but it was")
}

func (s *KeeperTestSuite) TestToggleOracle() {
	oracles := s.createOracles()
	oracleToToggle := oracles[1]

	// Set the oracle to inactive
	s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, false)
	oracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToToggle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().False(oracle.Active, "oracle should have been marked inactive")

	// Set it back to active
	s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, oracleToToggle.ChainId, true)
	oracle, found = s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracleToToggle.ChainId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().True(oracle.Active, "oracle should have been marked as active")
}

func (s *KeeperTestSuite) TestGetOracleFromConnectionId() {
	oracles := s.createOracles()

	// Get oracle using connection Id
	expectedOracle := oracles[1]
	actualOracle, found := s.App.ICAOracleKeeper.GetOracleFromConnectionId(s.Ctx, expectedOracle.ConnectionId)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().Equal(expectedOracle, actualOracle)

	// Attempt to get an oracle with a fake connectionId - should fail
	actualOracle, found = s.App.ICAOracleKeeper.GetOracleFromConnectionId(s.Ctx, "fake-connection-id")
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
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, transfertypes.PortID, ibctesting.FirstChannelID)
	s.Require().True(found, "transfer channel should have been found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, ibctesting.FirstChannelID, channel)

	// Try again, it should be false this time
	isOpen = s.App.ICAOracleKeeper.IsOracleICAChannelOpen(s.Ctx, oracle)
	s.Require().False(isOpen, "channel should now be closed")

	// Try with a fake channel
	oracle.ChannelId = "fake_channel"
	isOpen = s.App.ICAOracleKeeper.IsOracleICAChannelOpen(s.Ctx, oracle)
	s.Require().False(isOpen, "channel does not exist")
}
