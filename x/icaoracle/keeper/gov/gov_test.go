package gov_test

import (
	"strconv"
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v11/app/apptesting"
	"github.com/Stride-Labs/stride/v11/x/icaoracle/keeper/gov"
	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"
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

// Helper function to create 5 oracle objects with various attributes
func (s *KeeperTestSuite) CreateTestOracles() []types.Oracle {
	oracles := []types.Oracle{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)

		channelId := "channel-" + suffix
		portId := "port-" + suffix

		oracle := types.Oracle{
			ChainId:         "chain-" + suffix,
			ConnectionId:    "connection-" + suffix,
			ChannelId:       channelId,
			PortId:          portId,
			IcaAddress:      "oracle-address",
			ContractAddress: "contract-address",
			Active:          true,
		}

		oracles = append(oracles, oracle)
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

		// Create open ICA channel
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, channelId, channeltypes.Channel{
			State: channeltypes.OPEN,
		})
	}
	return oracles
}

func (s *KeeperTestSuite) TestGovToggleOracle() {
	oracles := s.CreateTestOracles()

	oracleIndexToToggle := 1
	oracleToToggle := oracles[oracleIndexToToggle]

	// Set the oracle to inactive
	err := gov.ToggleOracle(s.Ctx, s.App.ICAOracleKeeper, &types.ToggleOracleProposal{
		OracleChainId: oracleToToggle.ChainId,
		Active:        false,
	})
	s.Require().NoError(err)

	// Confirm it's the only oracle inactive
	for i, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracle.ChainId)
		s.Require().True(found, "oracle %s does not exist", oracle.ChainId)

		if i == oracleIndexToToggle {
			s.Require().False(oracle.Active, "oracle %s should have been toggled to inactive", oracle.ChainId)
		} else {
			s.Require().True(oracle.Active, "oracle %s should still be active", oracle.ChainId)
		}
	}

	// Set it back to active
	err = gov.ToggleOracle(s.Ctx, s.App.ICAOracleKeeper, &types.ToggleOracleProposal{
		OracleChainId: oracleToToggle.ChainId,
		Active:        true,
	})
	s.Require().NoError(err)

	// Confirm all oracles are active again
	for _, oracle := range s.App.ICAOracleKeeper.GetAllOracles(s.Ctx) {
		_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracle.ChainId)
		s.Require().True(found, "oracle %s does not exist", oracle.ChainId)
		s.Require().True(oracle.Active, "oracle %s should still be active", oracle.ChainId)
	}
}

func (s *KeeperTestSuite) TestGovRemoveOracle() {
	oracles := s.CreateTestOracles()

	oracleIndexToRemove := 1
	oracleToRemove := oracles[oracleIndexToRemove]

	// Remove the oracle thorugh goverance
	err := gov.RemoveOracle(s.Ctx, s.App.ICAOracleKeeper, &types.RemoveOracleProposal{
		OracleChainId: oracleToRemove.ChainId,
	})
	s.Require().NoError(err)

	// Confirm only one oracle was removed
	remainingOracles := s.App.ICAOracleKeeper.GetAllOracles(s.Ctx)
	s.Require().Len(remainingOracles, len(oracles)-1, "number of oracles after removal")

	// Confirm the other oracles are still there
	for i, oracle := range oracles {
		_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, oracle.ChainId)
		if i == oracleIndexToRemove {
			s.Require().False(found, "oracle %s should have been removed", oracle.ChainId)
		} else {
			s.Require().True(found, "oracle %s should not have been removed", oracle.ChainId)
		}
	}
}
