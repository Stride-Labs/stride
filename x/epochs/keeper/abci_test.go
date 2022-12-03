package keeper_test

import (
	"fmt"
	"time"

	"github.com/Stride-Labs/stride/v4/x/epochs"
	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

func (suite *KeeperTestSuite) TestEpochInfoChangesBeginBlockerAndInitGenesis() {
	var epochInfo types.EpochInfo

	now := time.Now()

	testCases := []struct {
		expCurrentEpochStartTime   time.Time
		expCurrentEpochStartHeight int64
		expCurrentEpoch            int64
		expInitialEpochStartTime   time.Time
		fn                         func()
	}{
		{
			// Only advance 2 seconds, do not increment epoch
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now,
			fn: func() {
				ctx := suite.Ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartHeight: 2,
			expCurrentEpochStartTime:   now,
			expCurrentEpoch:            1,
			expInitialEpochStartTime:   now,
			fn: func() {
				ctx := suite.Ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 31))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		// Test that incrementing _exactly_ 1 month increments the epoch count.
		{
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				ctx := suite.Ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				ctx := suite.Ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expCurrentEpochStartHeight: 3,
			expCurrentEpochStartTime:   now.Add(time.Hour * 24 * 31),
			expCurrentEpoch:            2,
			expInitialEpochStartTime:   now,
			fn: func() {
				ctx := suite.Ctx.WithBlockHeight(2).WithBlockTime(now.Add(time.Second))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(time.Hour * 24 * 32))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				ctx.WithBlockHeight(4).WithBlockTime(now.Add(time.Hour * 24 * 33))
				suite.App.EpochsKeeper.BeginBlocker(ctx)
				epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			suite.SetupTest() // reset
			ctx := suite.Ctx

			// On init genesis, default epochs information is set
			// To check init genesis again, should make it fresh status
			epochInfos := suite.App.EpochsKeeper.AllEpochInfos(ctx)
			for _, epochInfo := range epochInfos {
				suite.App.EpochsKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
			}

			// check init genesis
			ctx = ctx.WithBlockHeight(1).WithBlockTime(now)
			epochs.InitGenesis(ctx, suite.App.EpochsKeeper, types.GenesisState{
				Epochs: []types.EpochInfo{
					{
						Identifier:              "monthly",
						StartTime:               time.Time{},
						Duration:                time.Hour * 24 * 31,
						CurrentEpoch:            0,
						CurrentEpochStartHeight: ctx.BlockHeight(),
						CurrentEpochStartTime:   time.Time{},
						EpochCountingStarted:    false,
					},
				},
			})

			tc.fn()

			suite.Require().Equal(epochInfo.Identifier, "monthly")
			suite.Require().Equal(epochInfo.StartTime.UTC().String(), tc.expInitialEpochStartTime.UTC().String())
			suite.Require().Equal(epochInfo.Duration, time.Hour*24*31)
			suite.Require().Equal(epochInfo.CurrentEpoch, tc.expCurrentEpoch)
			suite.Require().Equal(epochInfo.CurrentEpochStartHeight, tc.expCurrentEpochStartHeight)
			suite.Require().Equal(epochInfo.CurrentEpochStartTime.UTC().String(), tc.expCurrentEpochStartTime.UTC().String())
			suite.Require().Equal(epochInfo.EpochCountingStarted, true)
		})
	}
}

func (suite *KeeperTestSuite) TestEpochStartingOneMonthAfterInitGenesis() {
	ctx := suite.Ctx
	// On init genesis, default epochs information is set
	// To check init genesis again, should make it fresh status
	epochInfos := suite.App.EpochsKeeper.AllEpochInfos(ctx)
	for _, epochInfo := range epochInfos {
		suite.App.EpochsKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
	}

	now := time.Now()
	week := time.Hour * 24 * 7
	month := time.Hour * 24 * 30
	initialBlockHeight := int64(1)
	ctx = ctx.WithBlockHeight(initialBlockHeight).WithBlockTime(now)

	epochs.InitGenesis(ctx, suite.App.EpochsKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "monthly",
				StartTime:               now.Add(month),
				Duration:                time.Hour * 24 * 30,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: ctx.BlockHeight(),
				CurrentEpochStartTime:   time.Time{},
				EpochCountingStarted:    false,
			},
		},
	})

	// epoch not started yet
	epochInfo, _ := suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
	suite.Require().Equal(epochInfo.EpochCountingStarted, false)

	// after 1 week
	ctx = ctx.WithBlockHeight(2).WithBlockTime(now.Add(week))
	suite.App.EpochsKeeper.BeginBlocker(ctx)

	// epoch not started yet
	epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(0))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, initialBlockHeight)
	suite.Require().Equal(epochInfo.CurrentEpochStartTime, time.Time{})
	suite.Require().Equal(epochInfo.EpochCountingStarted, false)

	// after 1 month
	ctx = ctx.WithBlockHeight(3).WithBlockTime(now.Add(month))
	suite.App.EpochsKeeper.BeginBlocker(ctx)

	// epoch started
	epochInfo, _ = suite.App.EpochsKeeper.GetEpochInfo(ctx, "monthly")
	suite.Require().Equal(epochInfo.CurrentEpoch, int64(1))
	suite.Require().Equal(epochInfo.CurrentEpochStartHeight, ctx.BlockHeight())
	suite.Require().Equal(epochInfo.CurrentEpochStartTime.UTC().String(), now.Add(month).UTC().String())
	suite.Require().Equal(epochInfo.EpochCountingStarted, true)
}
