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

// Tests TransferRewardTokensHostToTrade and GetHostToTradeTransferMsg
func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade() {
	// Create an ICA channel for the transfer submission
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)
	channelId, portId := s.CreateICAChannel(owner)

	// Define components of transfer message
	hostToRewardChannelId := "channel-0"
	rewardToTradeChannelId := "channel-1"

	rewardDenomOnHostZone := "ibc/reward_on_host"
	rewardDenomOnRewardZone := "reward_on_reward"

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

		HostAccount: types.ICAAccount{
			Address:      withdrawalAddress,
			ConnectionId: ibctesting.FirstConnectionID,
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
	actualMsg, err := s.App.StakeibcKeeper.GetHostToTradeTransferMsg(s.Ctx, transferAmount, route)
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
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, minSwapAmount.Sub(sdkmath.OneInt()), route)
	s.Require().NoError(err, "no error expected when submitting transfer")

	endSequence := s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(sequenceAfterTransfer, endSequence, "sequence number should NOT have incremented")

	// Remove the connection ID so the ICA fails
	invalidRoute := route
	invalidRoute.HostAccount.ConnectionId = ""
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, minSwapAmount.Sub(sdkmath.OneInt()), invalidRoute)
	s.Require().NoError(err, "no error expected when submitting transfer")

	endSequence = s.MustGetNextSequenceNumber(portId, channelId)
	s.Require().Equal(sequenceAfterTransfer, endSequence, "sequence number should NOT have incremented")

	// Delete the epoch tracker and call each function, confirming they both fail
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err = s.App.StakeibcKeeper.GetHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
}
