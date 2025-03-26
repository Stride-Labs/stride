package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/x/epochs/types"
)

func (s *KeeperTestSuite) TestEpochLifeCycle() {
	s.SetupTest()

	epochInfo := types.EpochInfo{
		Identifier:            "monthly",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	s.App.EpochsKeeper.SetEpochInfo(s.Ctx, epochInfo)
	epochInfoSaved, _ := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, "monthly")
	s.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := s.App.EpochsKeeper.AllEpochInfos(s.Ctx)
	s.Require().Len(allEpochs, 6)
	s.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	s.Require().Equal(allEpochs[1].Identifier, "hour")
	s.Require().Equal(allEpochs[2].Identifier, "mint")
	s.Require().Equal(allEpochs[3].Identifier, "monthly")
	s.Require().Equal(allEpochs[4].Identifier, "stride_epoch")
	s.Require().Equal(allEpochs[5].Identifier, "week")
}
