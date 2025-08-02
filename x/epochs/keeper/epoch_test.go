package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v27/x/epochs/types"
)

func (s *KeeperTestSuite) TestEpochLifeCycle() {
	s.SetupTest()

	// Set default epochs
	for _, epochInfo := range types.DefaultGenesis().Epochs {
		s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochInfo)
	}

	// Add the month epoch
	epochInfo := types.EpochInfo{
		Identifier:            "monthly",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochInfo)

	// Confirm looking up the monthly epoch
	epochInfoSaved, _ := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "monthly")
	s.Require().Equal(epochInfo, epochInfoSaved)

	// Confirm looking up all epochs
	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().Len(allEpochs, 6)
	s.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	s.Require().Equal(allEpochs[1].Identifier, "hour")
	s.Require().Equal(allEpochs[2].Identifier, "mint")
	s.Require().Equal(allEpochs[3].Identifier, "monthly")
	s.Require().Equal(allEpochs[4].Identifier, "stride_epoch")
	s.Require().Equal(allEpochs[5].Identifier, "week")
}
