package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetHostToTradeTransferMsg() {
	hostToRewardChannelId := "channel-0"
	rewardToTradeChannelId := "channel-1"

	rewardDenomOnHostZone := "ibc/reward_on_host"
	rewardDenomOnRewardZone := "reward_on_reward"

	withdrawalAddress := "withdrawal_address"
	unwindAddress := "unwind_address"
	tradeAddress := "trade_address"

	transferAmount := sdk.NewInt(1000)
	transferToken := sdk.NewCoin(rewardDenomOnHostZone, transferAmount)

	currentTime := uint64(10_000_000_000)               // 10s in nanoseconds
	epochEndTime := uint64(20_000_000_000)              // 20s in nanoseconds
	transfer1TimeoutTimestamp := uint64(15_000_000_000) // halfway through
	transfer2TimeoutDuration := "5s"

	// Create a mock context to override the current time
	s.Ctx = s.Ctx.WithBlockTime(time.Unix(0, int64(currentTime)))

	// Create a trade route with the relevant addresses and transfer channels
	route := types.TradeRoute{
		HostToRewardChannelId:  hostToRewardChannelId,
		RewardToTradeChannelId: rewardToTradeChannelId,

		RewardDenomOnHostZone:   rewardDenomOnHostZone,
		RewardDenomOnRewardZone: rewardDenomOnRewardZone,

		HostAccount: types.ICAAccount{
			Address: withdrawalAddress,
		},
		RewardAccount: types.ICAAccount{
			Address: unwindAddress,
		},
		TradeAccount: types.ICAAccount{
			Address: tradeAddress,
		},
	}

	// Create an epoch tracker to dictate the timeout
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: epochEndTime,
		Duration:           epochEndTime - currentTime,
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
		TimeoutTimestamp: transfer1TimeoutTimestamp,
		Memo:             memoJSON,
	}

	// Confirm the generated message matches expectations
	actualMsg, err := s.App.StakeibcKeeper.GetHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().NoError(err, "no error expected when building transfer message")
	s.Require().Equal(expectedMsg, actualMsg, "transfer message should have matched")

	// Delete the epoch tracker so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err = s.App.StakeibcKeeper.GetHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
}
