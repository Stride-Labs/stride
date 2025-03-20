package keeper_test

import (
	"time"

	"github.com/Stride-Labs/stride/v26/x/epochs/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "A",
				StartTime:               time.Time{},
				Duration:                time.Hour * 24 * 7,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
			{
				Identifier:              "B",
				StartTime:               time.Time{},
				Duration:                time.Hour * 24 * 7,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: 0,
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
		},
	}

	s.App.EpochsKeeper.InitGenesis(s.Ctx, genesisState)
	exported := s.App.EpochsKeeper.ExportGenesis(s.Ctx)

	s.Require().Equal(genesisState, *exported)
}
