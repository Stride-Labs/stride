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
			Moniker:      "moniker-" + suffix,
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
	moniker := expectedOracle.Moniker

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, moniker)
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
	monikerToRemove := oracleToRemove.Moniker

	s.App.ICAOracleKeeper.RemoveOracle(s.Ctx, monikerToRemove)
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, monikerToRemove)
	s.Require().False(found, "the removed oracle should not have been found, but it was")
}

func (s *KeeperTestSuite) TestToggleOracle() {
	oracles := s.createOracles()

	oracleToToggle := oracles[1]
	monikerToToggle := oracleToToggle.Moniker

	s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, monikerToToggle, false)
	oracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, monikerToToggle)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().False(oracle.Active, "oracle should have been marked inactive")

	s.App.ICAOracleKeeper.ToggleOracle(s.Ctx, monikerToToggle, true)
	oracle, found = s.App.ICAOracleKeeper.GetOracle(s.Ctx, monikerToToggle)
	s.Require().True(found, "oracle should have been found, but was not")
	s.Require().True(oracle.Active, "oracle should have been marked as active")
}
