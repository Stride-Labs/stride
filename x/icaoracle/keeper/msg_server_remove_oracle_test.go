package keeper_test

import (
	"fmt"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

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
	_, err := s.GetMsgServer().RemoveOracle(s.Ctx, &types.MsgRemoveOracle{
		Authority:     s.App.ICAOracleKeeper.GetAuthority(),
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
