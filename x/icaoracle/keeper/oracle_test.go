package keeper_test

import (
	"strconv"

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
