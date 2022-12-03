package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()
	ctx := suite.Ctx

	epochInfo := types.EpochInfo{
		Identifier:            "monthly",
		StartTime:             time.Time{},
		Duration:              time.Hour * 24 * 30,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
	}
	suite.App.EpochsKeeper.SetEpochInfo(ctx, epochInfo)
	epochInfoSaved, _ := suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.App.EpochsKeeper.AllEpochInfos(ctx)
	suite.Require().Len(allEpochs, 5)
	suite.Require().Equal(allEpochs[0].Identifier, "day") // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "mint")
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
	suite.Require().Equal(allEpochs[3].Identifier, "stride_epoch")
	suite.Require().Equal(allEpochs[4].Identifier, "week")
}
