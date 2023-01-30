package keeper_test

import (
	"time"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()
	ctx := suite.Ctx

	epochInfo := types.EpochInfo{
		Identifier:              "monthly",
		StartTime:               time.Time{},
		Duration:                time.Hour * 24 * 30,
		CurrentEpoch:            sdk.ZeroInt(),
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
		CurrentEpochStartHeight: sdk.ZeroInt(),
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
