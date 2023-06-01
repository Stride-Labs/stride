package keeper_test

import (
	gocontext "context"
	"time"

	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/x/epochs/types"
)

func (suite *KeeperTestSuite) TestQueryEpochInfos() {
	suite.SetupTest()
	queryClient := suite.queryClient

	chainStartTime := suite.Ctx.BlockTime()

	// Invalid param
	epochInfosResponse, err := queryClient.EpochInfos(gocontext.Background(), &types.QueryEpochsInfoRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(epochInfosResponse.Epochs, 5)

	// check if EpochInfos are correct
	suite.Require().Equal(epochInfosResponse.Epochs[0], types.EpochInfo{
		Identifier:            "day",
		StartTime:             chainStartTime,
		Duration:              time.Hour * 24,
		CurrentEpoch:          int64(0),
		CurrentEpochStartTime: chainStartTime,
		EpochCountingStarted:  false,
	})

	suite.Require().Equal(epochInfosResponse.Epochs[1], types.EpochInfo{
		Identifier:            "hour",
		StartTime:             chainStartTime,
		Duration:              time.Hour,
		CurrentEpoch:          int64(0),
		CurrentEpochStartTime: chainStartTime,
		EpochCountingStarted:  false,
	})

	suite.Require().Equal(epochInfosResponse.Epochs[2], types.EpochInfo{
		Identifier:            "mint",
		StartTime:             chainStartTime,
		Duration:              time.Minute * 60,
		CurrentEpoch:          int64(0),
		CurrentEpochStartTime: chainStartTime,
		EpochCountingStarted:  false,
	})

	suite.Require().Equal(epochInfosResponse.Epochs[3], types.EpochInfo{
		Identifier:            "stride_epoch",
		StartTime:             chainStartTime,
		Duration:              time.Hour * 6,
		CurrentEpoch:          int64(0),
		CurrentEpochStartTime: chainStartTime,
		EpochCountingStarted:  false,
	})

	suite.Require().Equal(epochInfosResponse.Epochs[4], types.EpochInfo{
		Identifier:            "week",
		StartTime:             chainStartTime,
		Duration:              time.Hour * 24 * 7,
		CurrentEpoch:          int64(0),
		CurrentEpochStartTime: chainStartTime,
		EpochCountingStarted:  false,
	})
}
