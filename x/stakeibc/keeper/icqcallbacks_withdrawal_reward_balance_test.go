package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v27/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v27/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

// WithdrawalRewardBalanceCallback will trigger TransferRewardTokensHostToTrade
// Therefore we need to setup traderoute fields used in the entire transfer (with pfm)
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceCallbackTestCase() BalanceQueryCallbackTestCase {
	// Create the connection between Stride and HostChain with the withdrawal account initialized
	withdrawalAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)
	withdrawalChannelId, withdrawalPortId := s.CreateICAChannel(withdrawalAccountOwner)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      HostChainId,
		HostDenom:    HostDenom,
		ConnectionId: ibctesting.FirstConnectionID,
	})

	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardDenomOnHostZone:   "ibc/reward_on_host",

		HostToRewardChannelId:  "channel-2",
		RewardToTradeChannelId: "channel-3",

		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      "withdrawal-address",
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_WITHDRAWAL,
		},
		RewardAccount: types.ICAAccount{
			Address: "reward-address",
		},
		TradeAccount: types.ICAAccount{
			Address: "trade-address",
		},

		MinTransferAmount: sdk.ZeroInt(),
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// Create and set the epoch tracker for timeouts
	timeoutDuration := time.Second * 30
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, timeoutDuration)

	// Build query object and serialized query response
	balance := sdkmath.NewInt(1_000_000)
	callbackDataBz, _ := proto.Marshal(&types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	})
	query := icqtypes.Query{
		ChainId:      HostChainId,
		CallbackData: callbackDataBz,
	}
	queryResponse := s.CreateBalanceQueryResponse(balance.Int64(), route.RewardDenomOnHostZone)

	return BalanceQueryCallbackTestCase{
		TradeRoute: route,
		Balance:    balance,
		Response: ICQCallbackArgs{
			Query:        query,
			CallbackArgs: queryResponse,
		},
		ChannelID: withdrawalChannelId,
		PortID:    withdrawalPortId,
	}
}

// Verify that a normal WithdrawalRewardBalanceCallback does fire off the ICA for transfer
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_Successful() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// ICA inside of TransferRewardTokensHostToTrade should execute but it uses submitTXWithoutCallback
	// So no need to confirm ICA callback data was stored and no need to confirm callback args values

	// Confirm ICA was submitted by checking that the sequence number incremented
	s.CheckICATxSubmitted(tc.PortID, tc.ChannelID, func() error {
		return keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	})
}

// Verify that if the amount returned by the ICQ response is less than the min_swap_amount, no transfer happens
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_SuccessfulNoTransfer() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Set min transfer amount to be greater than the transfer amount
	route := tc.TradeRoute
	route.MinTransferAmount = tc.Balance.Add(sdkmath.OneInt())
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// ICA inside of TransferRewardTokensHostToTrade should not actually execute because of min_swap_amount
	s.CheckICATxNotSubmitted(tc.PortID, tc.ChannelID, func() error {
		return keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	})
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_ZeroBalance() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Replace the query response with a coin that has a zero amount
	tc.Response.CallbackArgs = s.CreateBalanceQueryResponse(0, tc.TradeRoute.RewardDenomOnHostZone)

	// Confirm the transfer ICA was never sent
	s.CheckICATxNotSubmitted(tc.PortID, tc.ChannelID, func() error {
		return keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	})
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_InvalidArgs() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "unable to determine balance from query response")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_InvalidCallbackData() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Update the callback data so that it can't be successfully unmarshalled
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = []byte("random bytes")

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to unmarshal trade reward balance callback data")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_TradeRouteNotFound() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Update the callback data so that it keys to a trade route that doesn't exist
	invalidCallbackDataBz, _ := proto.Marshal(&types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   "different-host-denom",
	})
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = invalidCallbackDataBz

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "trade route not found")
}

func (s *KeeperTestSuite) TestWithdrawalRewardBalanceCallback_FailedSubmitTx() {
	tc := s.SetupWithdrawalRewardBalanceCallbackTestCase()

	// Remove connectionId from host ICAAccount on TradeRoute so the ICA tx fails
	invalidRoute := tc.TradeRoute
	invalidRoute.HostAccount.ConnectionId = "bad-connection"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, invalidRoute)

	err := keeper.WithdrawalRewardBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")
}
