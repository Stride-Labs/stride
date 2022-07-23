package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.EpochInfo{
		Identifier:            "monthly",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	suite.App.EpochsKeeper.SetEpochInfo(suite.Ctx, epochInfo)
	epochInfoSaved, _ := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, "monthly")
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.App.EpochsKeeper.AllEpochInfos(suite.Ctx)
	suite.Require().Len(allEpochs, 4)
	suite.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "monthly")
	suite.Require().Equal(allEpochs[2].Identifier, "stride_epoch")
	suite.Require().Equal(allEpochs[3].Identifier, "week")
}

// EPOCHS
