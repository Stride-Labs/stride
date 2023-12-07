package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupTradeConvertedBalanceCallbackTestCase() BalanceQueryCallbackTestCase {
	// Create the connection between Stride and HostChain with the withdrawal account initialized
	// TODO [DYDX]: Replace with trade route formatter
	tradeAccountOwner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_CONVERTER_TRADE)
	tradeChannelId, tradePortId := s.CreateICAChannel(tradeAccountOwner)

	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		HostDenomOnTradeZone:    "ibc/host_on_trade",

		TradeToHostChannelId: "channel-2",

		HostAccount: types.ICAAccount{
			Address: "withdrawal-address",
		},
		TradeAccount: types.ICAAccount{
			Address:      "trade-address",
			ConnectionId: ibctesting.FirstConnectionID,
		},
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
	query := icqtypes.Query{CallbackData: callbackDataBz}
	queryResponse := s.CreateBalanceQueryResponse(balance.Int64(), route.HostDenomOnTradeZone)

	// Get start sequence to test ICA submission
	startSequence := s.MustGetNextSequenceNumber(tradePortId, tradeChannelId)

	return BalanceQueryCallbackTestCase{
		TradeRoute: route,
		Balance:    balance,
		Response: ICQCallbackArgs{
			Query:        query,
			CallbackArgs: queryResponse,
		},
		ChannelID:     tradeChannelId,
		PortID:        tradePortId,
		StartSequence: startSequence,
	}
}

// Verify that a normal TradeConvertedBalanceCallback does fire off the ICA for transfer
func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_Successful() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().NoError(err)

	// Confirm the sequence number was incremented
	endSequence := s.MustGetNextSequenceNumber(tc.PortID, tc.ChannelID)
	s.Require().Equal(endSequence, tc.StartSequence+1, "sequence number should increase after callback executed")
}

func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_ZeroBalance() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	// Replace the query response with a coin that has a zero amount
	tc.Response.CallbackArgs = s.CreateBalanceQueryResponse(0, tc.TradeRoute.HostDenomOnHostZone)

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().NoError(err)

	// Confirm the sequence number was NOT incremented, meaning the transfer ICA was not called
	endSequence := s.MustGetNextSequenceNumber(tc.PortID, tc.ChannelID)
	s.Require().Equal(endSequence, tc.StartSequence, "sequence number should NOT have increased, no transfer should happen")
}

func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_InvalidArgs() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "unable to determine balance from query response")
}

func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_InvalidCallbackData() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	// Update the callback data so that it can't be successfully unmarshalled
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = []byte("random bytes")

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to unmarshal trade reward balance callback data")
}

func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_TradeRouteNotFound() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	// Update the callback data so that it keys to a trade route that doesn't exist
	invalidCallbackDataBz, _ := proto.Marshal(&types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   "different-host-denom",
	})
	invalidQuery := tc.Response.Query
	invalidQuery.CallbackData = invalidCallbackDataBz

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "trade route not found")
}

func (s *KeeperTestSuite) TestTradeConvertedBalanceCallback_FailedSubmitTx() {
	tc := s.SetupTradeConvertedBalanceCallbackTestCase()

	// Remove connectionId from host ICAAccount on TradeRoute so the ICA tx fails
	invalidRoute := tc.TradeRoute
	invalidRoute.TradeAccount.ConnectionId = "bad-connection"
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, invalidRoute)

	err := keeper.TradeConvertedBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.Response.CallbackArgs, tc.Response.Query)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")
}
