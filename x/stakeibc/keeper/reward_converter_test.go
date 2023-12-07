package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// Tests TransferRewardTokensHostToTrade and BuildHostToTradeTransferMsg
func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade() {
	// Create an ICA channel for the transfer submission
	owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, types.ICAAccountType_WITHDRAWAL)
	channelId, portId := s.CreateICAChannel(owner)

	// Define components of transfer message
	hostToRewardChannelId := "channel-0"
	rewardToTradeChannelId := "channel-1"

	rewardDenomOnHostZone := "ibc/reward_on_host"
	rewardDenomOnRewardZone := RewardDenom

	withdrawalAddress := "withdrawal_address"
	unwindAddress := "unwind_address"
	tradeAddress := "trade_address"

	transferAmount := sdk.NewInt(1000)
	transferToken := sdk.NewCoin(rewardDenomOnHostZone, transferAmount)
	minSwapAmount := sdk.NewInt(500)

	currentTime := s.Ctx.BlockTime()
	epochLength := time.Second * 10                               // 10 seconds
	epochEndTime := currentTime.Add(time.Second * 10)             // 10 seconds from now
	transfer1TimeoutTimestamp := currentTime.Add(time.Second * 5) // 5 seconds from now (halfway through)
	transfer2TimeoutDuration := "5s"

	// Create a trade route with the relevant addresses and transfer channels
	route := types.TradeRoute{
		HostToRewardChannelId:  hostToRewardChannelId,
		RewardToTradeChannelId: rewardToTradeChannelId,

		RewardDenomOnHostZone:   rewardDenomOnHostZone,
		RewardDenomOnRewardZone: rewardDenomOnRewardZone,
		HostDenomOnHostZone:     HostDenom,

		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      withdrawalAddress,
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_WITHDRAWAL,
		},
		RewardAccount: types.ICAAccount{
			Address: unwindAddress,
		},
		TradeAccount: types.ICAAccount{
			Address: tradeAddress,
		},

		TradeConfig: types.TradeConfig{
			MinSwapAmount: minSwapAmount,
		},
	}

	// Create an epoch tracker to dictate the timeout
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: uint64(epochEndTime.UnixNano()),
		Duration:           uint64(epochLength.Nanoseconds()),
	})

	// Define the expected transfer message using all the above
	memoJSON := fmt.Sprintf(`{"forward":{"receiver":"%s","port":"transfer","channel":"%s","timeout":"%s","retries":0}}`,
		tradeAddress, rewardToTradeChannelId, transfer2TimeoutDuration)

	expectedMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostToRewardChannelId,
		Token:            transferToken,
		Sender:           withdrawalAddress,
		Receiver:         unwindAddress,
		TimeoutTimestamp: uint64(transfer1TimeoutTimestamp.UnixNano()),
		Memo:             memoJSON,
	}

	// Confirm the generated message matches expectations
	actualMsg, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().NoError(err, "no error expected when building transfer message")
	s.Require().Equal(expectedMsg, actualMsg, "transfer message should have matched")

	// Call the main transfer function and confirm the sequence number increments
	startSequence := s.MustGetNextSequenceNumber(portId, channelId)

	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, transferAmount, route)
	s.Require().NoError(err, "no error expected when submitting transfer")

	sequenceAfterTransfer := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(startSequence+1, sequenceAfterTransfer, "sequence number should have incremented")

	// Attempt to call the function again with an transfer amount below the min,
	// it should not submit an ICA
	invalidTransferAmount := minSwapAmount.Sub(sdkmath.OneInt())
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, invalidTransferAmount, route)
	s.Require().NoError(err, "no error expected when submitting transfer with amount below minimum")

	endSequence := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(sequenceAfterTransfer, endSequence, "sequence number should NOT have incremented")

	// Remove the connection ID and confirm the ICA fails
	invalidRoute := route
	invalidRoute.HostAccount.ConnectionId = ""
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")

	// Check unregisted ICA addresses cause failures
	invalidRoute = route
	invalidRoute.HostAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no host account found")

	invalidRoute = route
	invalidRoute.RewardAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no reward account found")

	invalidRoute = route
	invalidRoute.TradeAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")

	// Delete the epoch tracker and call each function, confirming they both fail
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
}

func (s *KeeperTestSuite) TestTransferConvertedTokensTradeToHost() {
	transferAmount := sdkmath.NewInt(1000)

	// Register a trade ICA account for the transfer
	channelId, portId := s.CreateICAChannel(types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_CONVERTER_TRADE))

	// Create trade route with fields needed for transfer
	route := types.TradeRoute{
		HostDenomOnTradeZone: HostDenom,
		TradeToHostChannelId: "channel-1",
		HostAccount: types.ICAAccount{
			Address: "host_address",
		},
		TradeAccount: types.ICAAccount{
			Address:      "trade_address",
			ConnectionId: ibctesting.FirstConnectionID,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// Create epoch tracker to dictate timeout
	s.CreateStrideEpochForICATimeout(time.Second * 10)

	// Confirm the sequence number was incremented after a successful send
	startSequence := s.MustGetNextSequenceNumber(portId, channelId)

	err := s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, route)
	s.Require().NoError(err, "no error expected when transfering tokens")

	endSequence := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(startSequence+1, endSequence, "sequence number should have incremented from transfer")

	// Attempt to send without a valid ICA address - it should fail
	invalidRoute := route
	invalidRoute.HostAccount.Address = ""
	err = s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no host account found")

	invalidRoute = route
	invalidRoute.TradeAccount.Address = ""
	err = s.App.StakeibcKeeper.TransferConvertedTokensTradeToHost(s.Ctx, transferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")
}

func (s *KeeperTestSuite) TestBuildSwapMsg() {
	poolId := uint64(100)
	tradeAddress := "trade_address"

	rewardDenom := "ibc/reward_on_trade"
	hostDenom := "ibc/host_on_trade"

	baseTradeRoute := types.TradeRoute{
		RewardDenomOnTradeZone: rewardDenom,
		HostDenomOnTradeZone:   hostDenom,

		TradeAccount: types.ICAAccount{
			Address: tradeAddress,
		},

		TradeConfig: types.TradeConfig{
			PoolId: poolId,
		},
	}

	testCases := []struct {
		name                string
		price               sdk.Dec
		maxAllowedSwapLoss  sdk.Dec
		minSwapAmount       sdkmath.Int
		maxSwapAmount       sdkmath.Int
		rewardAmount        sdkmath.Int
		expectedTradeAmount sdkmath.Int
		expectedMinOut      sdkmath.Int
		expectedError       string
	}{
		{
			// Reward Amount: 100, Min: 0, Max: 200 => Trade Amount: 100
			// Price: 1, Slippage: 5% => Min Out: 95
			name:               "swap 1",
			price:              sdk.MustNewDecFromStr("1.0"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.05"),

			maxSwapAmount:       sdkmath.NewInt(200),
			rewardAmount:        sdkmath.NewInt(100),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(95),
		},
		{
			// Reward Amount: 100, Min: 0, Max: 200 => Trade Amount: 100
			// Price: 0.70, Slippage: 10% => Min Out: 100 * 0.70 * 0.9 = 63
			name:               "swap 2",
			price:              sdk.MustNewDecFromStr("0.70"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.10"),

			maxSwapAmount:       sdkmath.NewInt(200),
			rewardAmount:        sdkmath.NewInt(100),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(63),
		},
		{
			// Reward Amount: 100, Min: 0, Max: 200 => Trade Amount: 100
			// Price: 1.80, Slippage: 15% => Min Out: 100 * 1.8 * 0.85 = 153
			name:               "swap 3",
			price:              sdk.MustNewDecFromStr("1.8"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.15"),

			maxSwapAmount:       sdkmath.NewInt(200),
			rewardAmount:        sdkmath.NewInt(100),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(153),
		},
		{
			// Reward Amount: 200, Min: 0, Max: 100 => Trade Amount: 100
			// Price: 1, Slippage: 5% => Min Out: 95
			name:               "capped by max swap amount",
			price:              sdk.MustNewDecFromStr("1.0"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.05"),

			maxSwapAmount:       sdkmath.NewInt(100),
			rewardAmount:        sdkmath.NewInt(200),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(95),
		},
		{
			// Reward Amount: 100, Min: 0, Max: 200 => Trade Amount: 100
			// Price: 1, Slippage: 5.001% => Min Out: 94.999 => truncated to 94
			name:               "int truncation in min out caused by decimal max swap allowed",
			price:              sdk.MustNewDecFromStr("1.0"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.05001"),

			maxSwapAmount:       sdkmath.NewInt(200),
			rewardAmount:        sdkmath.NewInt(100),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(94),
		},
		{
			// Reward Amount: 100, Min: 0, Max: 200 => Trade Amount: 100
			// Price: 0.9998, Slippage: 10% => Min Out: 89.991 => truncated to 89
			name:               "int truncation in min out caused by decimal price",
			price:              sdk.MustNewDecFromStr("0.9998"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.10"),

			maxSwapAmount:       sdkmath.NewInt(200),
			rewardAmount:        sdkmath.NewInt(100),
			expectedTradeAmount: sdkmath.NewInt(100),

			expectedMinOut: sdkmath.NewInt(89),
		},
		{
			// Reward Amount: 89234, Min: 0, Max: 23424 => Trade Amount: 23424
			// Price: 15.234323, Slippage: 9.234329%
			//   => Min Out: 23424 * 15.234323 * 0.90765671 = 323896.19 => truncates to 323896
			name:               "int truncation from random numbers",
			price:              sdk.MustNewDecFromStr("15.234323"),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0.09234329"),

			maxSwapAmount:       sdkmath.NewInt(23424),
			rewardAmount:        sdkmath.NewInt(89234),
			expectedTradeAmount: sdkmath.NewInt(23424),

			expectedMinOut: sdkmath.NewInt(323896),
		},
		{
			// Missing price
			name:               "missing price error",
			price:              sdk.ZeroDec(),
			maxAllowedSwapLoss: sdk.MustNewDecFromStr("0"),

			maxSwapAmount:       sdkmath.NewInt(0),
			rewardAmount:        sdkmath.NewInt(0),
			expectedTradeAmount: sdkmath.NewInt(0),
			expectedMinOut:      sdkmath.NewInt(0),

			expectedError: "Price not found for pool",
		},
	}

	for _, tc := range testCases {
		route := baseTradeRoute

		route.TradeConfig.SwapPrice = tc.price
		route.TradeConfig.MinSwapAmount = tc.minSwapAmount
		route.TradeConfig.MaxSwapAmount = tc.maxSwapAmount
		route.TradeConfig.MaxAllowedSwapLossRate = tc.maxAllowedSwapLoss

		msg, err := s.App.StakeibcKeeper.BuildSwapMsg(tc.rewardAmount, route)

		if tc.expectedError != "" {
			s.Require().ErrorContains(err, tc.expectedError, "%s - error expected", tc.name)
			continue
		}
		s.Require().Equal(tradeAddress, msg.Sender, "%s - sender", tc.name)
		s.Require().Equal(poolId, msg.Routes[0].PoolId, "%s - pool id", tc.name)

		s.Require().Equal(hostDenom, msg.Routes[0].TokenOutDenom, "%s - token out denom", tc.name)
		s.Require().Equal(rewardDenom, msg.TokenIn.Denom, "%s - token in denom", tc.name)

		s.Require().Equal(tc.expectedTradeAmount.Int64(), msg.TokenIn.Amount.Int64(), "%s - token in amount", tc.name)
		s.Require().Equal(tc.expectedMinOut.Int64(), msg.TokenOutMinAmount.Int64(), "%s - min token out", tc.name)
	}

	// Test with a missing ICA address
	invalidRoute := baseTradeRoute
	invalidRoute.TradeAccount.Address = ""
	_, err := s.App.StakeibcKeeper.BuildSwapMsg(sdk.NewInt(1), invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")
}

func (s *KeeperTestSuite) TestSwapRewardTokens() {
	// Create an ICA channel for the transfer submission
	owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	channelId, portId := s.CreateICAChannel(owner)

	minSwapAmount := sdkmath.NewInt(10)
	rewardAmount := sdkmath.NewInt(100)

	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,

		RewardDenomOnTradeZone: "ibc/reward_on_trade",
		HostDenomOnTradeZone:   "ibc/host_on_trade",

		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      "trade_address",
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
		},

		TradeConfig: types.TradeConfig{
			PoolId:                 100,
			SwapPrice:              sdk.OneDec(),
			MinSwapAmount:          minSwapAmount,
			MaxSwapAmount:          sdkmath.NewInt(1000),
			MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.1"),
		},
	}

	// Create an epoch tracker to dictate the timeout
	s.CreateEpochForICATimeout(epochtypes.HOUR_EPOCH, time.Minute) // arbitrary timeout time

	// Execute the swap and confirm the sequence number increments
	startSequence := s.MustGetNextSequenceNumber(portId, channelId)

	err := s.App.StakeibcKeeper.SwapRewardTokens(s.Ctx, rewardAmount, route)
	s.Require().NoError(err, "no error expected when submitting swap")

	sequenceAfterSwap := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(startSequence+1, sequenceAfterSwap, "sequence number should have incremented")

	// Attempt to call the function again with an swap amount below the min,
	// it should not submit an ICA
	invalidSwapAmount := minSwapAmount.Sub(sdkmath.OneInt())
	err = s.App.StakeibcKeeper.SwapRewardTokens(s.Ctx, invalidSwapAmount, route)
	s.Require().NoError(err, "no error expected when submitting transfer with amount below minimum")

	endSequence := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(sequenceAfterSwap, endSequence, "sequence number should NOT have incremented")

	// Remove the connection ID so the ICA fails
	invalidRoute := route
	invalidRoute.TradeAccount.ConnectionId = ""
	err = s.App.StakeibcKeeper.SwapRewardTokens(s.Ctx, rewardAmount, invalidRoute)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")

	// Delete the epoch tracker and confirm the swap fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.HOUR_EPOCH)
	err = s.App.StakeibcKeeper.SwapRewardTokens(s.Ctx, rewardAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
}
