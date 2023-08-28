package keeper_test

import (
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	oracle := types.Oracle{
		ChainId: "chain",
	}
	metric := types.Metric{
		Key:               "key",
		UpdateTime:        int64(1),
		DestinationOracle: "chain",
		Status:            types.MetricStatus_QUEUED,
	}

	genesisState := types.GenesisState{
		Params:  types.Params{},
		Oracles: []types.Oracle{oracle},
		Metrics: []types.Metric{metric},
	}

	s.App.ICAOracleKeeper.InitGenesis(s.Ctx, genesisState)
	exported := s.App.ICAOracleKeeper.ExportGenesis(s.Ctx)

	s.Require().Equal(genesisState, *exported)
}
