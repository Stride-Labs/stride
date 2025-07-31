package keeper_test

import (
	"time"

	"github.com/Stride-Labs/stride/v27/x/epochs/types"
)

func (s *KeeperTestSuite) TestGenesis() {
	genesisState := types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "A",
				StartTime:               s.Ctx.BlockTime(),
				Duration:                time.Hour * 24 * 7,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: s.Ctx.BlockHeight(),
				CurrentEpochStartTime:   s.Ctx.BlockTime(),
				EpochCountingStarted:    false,
			},
			{
				Identifier:              "B",
				StartTime:               s.Ctx.BlockTime(),
				Duration:                time.Hour * 24 * 7,
				CurrentEpoch:            1,
				CurrentEpochStartHeight: s.Ctx.BlockHeight(),
				CurrentEpochStartTime:   s.Ctx.BlockTime(),
				EpochCountingStarted:    false,
			},
		},
	}

	// Clear all epochs registered by default
	for _, epochInfo := range s.App.EpochsKeeper.AllEpochInfos(s.Ctx) {
		s.App.EpochsKeeper.DeleteEpochInfo(s.Ctx, epochInfo.Identifier)
	}

	s.App.EpochsKeeper.InitGenesis(s.Ctx, genesisState)
	exported := s.App.EpochsKeeper.ExportGenesis(s.Ctx)

	s.Require().Equal(genesisState, *exported)
}
