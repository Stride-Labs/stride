package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// WithdrawalRewardBalanceCallback will trigger TransferRewardTokensHostToTrade
// Therefore we need to setup traderoute fields used in the entire transfer (with pfm)
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceCallbackTestCase() BalanceQueryCallbackTestCase {
	// Create the connection between Stride and HostChain with the withdrawal account initialized
	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, types.ICAAccountType_WITHDRAWAL.String())
	withdrawalChannelId, withdrawalPortId := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]	
	withdrawalConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, withdrawalPortId, withdrawalChannelId)	

	hostICA := types.ICAAccount{
		ChainId: HostChainId,
		Type: types.ICAAccountType_WITHDRAWAL,
		ConnectionId: withdrawalConnectionId,
		Address: withdrawalAddress,
	}
	rewardICA := types.ICAAccount{
		ChainId: "noble-01",
		Type: types.ICAAccountType_CONVERTER_UNWIND,
		Address: HostICAAddress, // doesn't have to exist or be connected, but must be betch32 decode-able
	}
	tradeICA := types.ICAAccount{
		ChainId: OsmoChainId,
		Type: types.ICAAccountType_CONVERTER_TRADE,
		Address: HostICAAddress, // doesn't have to exist or be connected, but must be betch32 decode-able		
	}

	// Must initialize these or they will serialize differently from the default values
	tradeConfig := types.TradeConfig{
		SwapPrice: 				sdk.OneDec(),
		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.05"),
		MinSwapAmount: 			sdk.ZeroInt(),
		MaxSwapAmount: 			sdk.NewIntFromUint64(uint64(1_000_000)),
	}

	// Create and set the trade route for testing
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone: 	"ibc/reward_on_host",
		RewardDenomOnRewardZone:"reward",
		RewardDenomOnTradeZone: "ibc/reward_on_host",
		HostAccount: 			hostICA,
		RewardAccount: 			rewardICA,
		TradeAccount: 			tradeICA,
		HostToRewardChannelId:  "channel-02", //doesn't have to be real, has to exist
		RewardToTradeChannelId: "channel-17", //doesn't have to be real, has to exist
		TradeConfig: 			tradeConfig,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts
	timeoutDuration := time.Second * 30
	epochEndTime := uint64(s.Ctx.BlockTime().Add(timeoutDuration).UnixNano())
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: epochEndTime,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	callbackDataBz, _ := proto.Marshal(&tradeRoute)

	queryResponse := s.CreateBalanceQueryResponse(int64(1_000_000), tradeRoute.RewardDenomOnHostZone)

	return BalanceQueryCallbackTestCase{
		TradeRoute: tradeRoute,
		Response: ICQCallbackArgs{
			Query: icqtypes.Query{
				Id:      "0",
				ChainId: tradeRoute.HostAccount.ChainId,
				CallbackData: callbackDataBz,
			},
			CallbackArgs: queryResponse,
		},
		ChannelID: withdrawalChannelId,
		PortID: withdrawalPortId,
	}
}

// Verify that a normal WithdrawalRewardBalanceCallback does fire off the ICA for transfer
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_Successful() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, testCase.Response.CallbackArgs, testCase.Response.Query)
	s.Require().NoError(err)

	// ICA inside of TransferRewardTokensHostToTrade should execute but it uses submitTXWithoutCallback
	// So no need to confirm ICA callback data was stored and no need to confirm callback args values

	// Confirm the sequence number was incremented
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence+1, "sequence number should increase after callback executed")
}

// Verify that if the amount returned by the ICQ response is less than the min_swap_amount, no transfer happens
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_SuccessfulNoTransfer() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// The testCase.Response.CallbackArgs contains returned amount 1_000_000 so set min_swap_amount to be greater
	var tradeRoute types.TradeRoute
	tradeRoute.Unmarshal(testCase.Response.Query.CallbackData)
	tradeRoute.TradeConfig.MinSwapAmount = sdk.NewIntFromUint64(uint64(1_000_000_000))
	callbackDataBz, _ := proto.Marshal(&tradeRoute)
	testCase.Response.Query.CallbackData = callbackDataBz

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, testCase.Response.CallbackArgs, testCase.Response.Query)
	s.Require().NoError(err)

	// ICA inside of TransferRewardTokensHostToTrade should not actually execute because of min_swap_amount

	// Confirm the sequence number was NOT incremented
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence, "sequence number should NOT have increased, no transfer should happen")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_EmptyCallbackArgs() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")


	// Replace the query response an empty byte array (this happens when the account has not been registered yet)
	emptyCallbackArgs := []byte{}

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, emptyCallbackArgs, testCase.Response.Query)
	s.Require().NoError(err)

	// Confirm the sequence number was NOT incremented, meaning the transfer ICA was not called
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence, "sequence number should NOT have increased, no transfer should happen")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_ZeroBalance() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")

	// Replace the query response with a coin that has a zero amount
	testCase.Response.CallbackArgs = s.CreateBalanceQueryResponse(0, testCase.TradeRoute.RewardDenomOnHostZone)

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, testCase.Response.CallbackArgs, testCase.Response.Query)
	s.Require().NoError(err)

	// Confirm the sequence number was NOT incremented, meaning the transfer ICA was not called
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence, "sequence number should NOT have increased, no transfer should happen")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_ZeroBalanceImplied() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")

	// Replace the query response with a coin that has a nil amount
	coin := sdk.Coin{}
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	testCase.Response.CallbackArgs = coinBz

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, testCase.Response.CallbackArgs, testCase.Response.Query)
	s.Require().NoError(err)

	// Confirm the sequence number was NOT incremented, meaning the transfer ICA was not called
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence, "sequence number should NOT have increased, no transfer should happen")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_InvalidArgs() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, testCase.Response.Query)
	s.Require().ErrorContains(err, "unable to determine balance from query response")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_FailedSubmitTx() {
	testCase := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	startSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found before callback executed")

	// Remove connectionId from host ICAAccount on TradeRoute so the ICA tx fails
	testCase.TradeRoute.HostAccount.ConnectionId = "bad-connection"
	callbackDataBz, _ := proto.Marshal(&testCase.TradeRoute)
	testCase.Response.Query.CallbackData = callbackDataBz

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, testCase.Response.CallbackArgs, testCase.Response.Query)
	//s.Require().ErrorContains(err, "Failed to submit ICA tx")
	//s.Require().ErrorContains(err, "connection not found")	
	s.Require().NoError(err)

	// Confirm the sequence number was NOT incremented, meaning the transfer ICA was not called
	// Normally this would cause an error from the ICA tx send, but we consume those in the ApplyIfNoError wrapper
	endSequence, found := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, testCase.PortID, testCase.ChannelID)
	s.Require().True(found, "sequence number not found after callback should have executed ICA")
	s.Require().Equal(endSequence, startSequence, "sequence number should NOT have increased, no transfer should happen")	
}
