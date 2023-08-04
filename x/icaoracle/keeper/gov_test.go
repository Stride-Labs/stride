package keeper_test

import (
	"fmt"

	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestGovToggleOracle() {
	oracles := s.CreateTestOracles()

	oracleIndexToToggle := 1
	oracleToToggle := oracles[oracleIndexToToggle]

	// Set the oracle to inactive
	err := s.App.ICAOracleKeeper.HandleToggleOracleProposal(s.Ctx, &types.ToggleOracleProposal{
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
	err = s.App.ICAOracleKeeper.HandleToggleOracleProposal(s.Ctx, &types.ToggleOracleProposal{
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

	// Add metrics to that oracle
	for i := 0; i < 3; i++ {
		metric := types.Metric{
			Key:               fmt.Sprintf("key-%d", i),
			Value:             fmt.Sprintf("value-%d", i),
			BlockHeight:       s.Ctx.BlockHeight(),
			UpdateTime:        s.Ctx.BlockTime().Unix(),
			DestinationOracle: oracleToRemove.ChainId,
			Status:            types.MetricStatus_QUEUED,
		}
		s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)
	}

	// Remove the oracle thorugh goverance
	err := s.App.ICAOracleKeeper.HandleRemoveOracleProposal(s.Ctx, &types.RemoveOracleProposal{
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

	// Confirm the metrics were removed
	s.Require().Empty(s.App.ICAOracleKeeper.GetAllMetrics(s.Ctx), "all metrics removed")
}
