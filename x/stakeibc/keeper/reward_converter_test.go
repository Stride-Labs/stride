package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// Useful across all balance query icqcallback tests
type BalanceQueryCallbackTestCase struct {
	TradeRoute types.TradeRoute
	Response   ICQCallbackArgs
	Balance    sdkmath.Int
	ChannelID  string
	PortID     string
}

type TransferRewardHostToTradeTestCase struct {
	TradeRoute          types.TradeRoute
	TransferAmount      sdkmath.Int
	ExpectedTransferMsg transfertypes.MsgTransfer
	ChannelID           string
	PortID              string
}

// --------------------------------------------------------------
//                   Transfer Host to Trade
// --------------------------------------------------------------

func (s *KeeperTestSuite) SetupTransferRewardTokensHostToTradeTestCase() TransferRewardHostToTradeTestCase {
	// Create an ICA channel for the transfer submission
	owner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)
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
			SwapPrice:     sdk.OneDec(),
			MinSwapAmount: minSwapAmount,
		},
	}

	// Create an epoch tracker to dictate the timeout
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, epochLength)

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

	return TransferRewardHostToTradeTestCase{
		TradeRoute:          route,
		TransferAmount:      transferAmount,
		ExpectedTransferMsg: expectedMsg,
		ChannelID:           channelId,
		PortID:              portId,
	}
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_Success() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Confirm the generated message matches expectations
	actualMsg, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().NoError(err, "no error expected when building transfer message")
	s.Require().Equal(tc.ExpectedTransferMsg, actualMsg, "transfer message should have matched")
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_InvalidICAAddress() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Check unregisted ICA addresses cause failures
	invalidRoute := tc.TradeRoute
	invalidRoute.HostAccount.Address = ""
	_, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no host account found")

	invalidRoute = tc.TradeRoute
	invalidRoute.RewardAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no reward account found")

	invalidRoute = tc.TradeRoute
	invalidRoute.TradeAccount.Address = ""
	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "no trade account found")
}

func (s *KeeperTestSuite) TestBuildHostToTradeTransferMsg_EpochNotFound() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Delete the epoch tracker and confirm the message cannot be built
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().ErrorContains(err, "epoch not found")
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_Success() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Check that the transfer ICA is submitted when the function is called
	s.CheckICATxSubmitted(tc.PortID, tc.ChannelID, func() error {
		return s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	})
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_TransferAmountBelowMin() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Attempt to call the function with an transfer amount below the min,
	// it should not submit an ICA
	invalidTransferAmount := tc.TradeRoute.TradeConfig.MinSwapAmount.Sub(sdkmath.OneInt())
	s.CheckICATxNotSubmitted(tc.PortID, tc.ChannelID, func() error {
		return s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, invalidTransferAmount, tc.TradeRoute)
	})
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_NoPoolPrice() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Attempt to call the function with a route that does not have a pool price
	// If should not initiate the transfer (since the swap would be unable to execute)
	invalidRoute := tc.TradeRoute
	invalidRoute.TradeConfig.SwapPrice = sdk.ZeroDec()

	s.CheckICATxNotSubmitted(tc.PortID, tc.ChannelID, func() error {
		return s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, invalidRoute)
	})
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_FailedToSubmitICA() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Remove the connection ID and confirm the ICA submission fails
	invalidRoute := tc.TradeRoute
	invalidRoute.HostAccount.ConnectionId = ""

	err := s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, invalidRoute)
	s.Require().ErrorContains(err, "Failed to submit ICA tx")
}

func (s *KeeperTestSuite) TestTransferRewardTokensHostToTrade_EpochNotFound() {
	tc := s.SetupTransferRewardTokensHostToTradeTestCase()

	// Delete the epoch tracker and confirm the transfer cannot be initiated
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, tc.TransferAmount, tc.TradeRoute)
	s.Require().ErrorContains(err, "epoch not found")
}

// --------------------------------------------------------------
//                   Transfer Trade to Trade
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestTransferConvertedTokensTradeToHost() {
	transferAmount := sdkmath.NewInt(1000)

	// Register a trade ICA account for the transfer
	owner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	channelId, portId := s.CreateICAChannel(owner)

	// Create trade route with fields needed for transfer
	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,

		HostDenomOnTradeZone: "ibc/host-on-trade",
		TradeToHostChannelId: "channel-1",
		HostAccount: types.ICAAccount{
			Address: "host_address",
		},
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Address:      "trade_address",
			ConnectionId: ibctesting.FirstConnectionID,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, route)

	// Create epoch tracker to dictate timeout
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, time.Second*10)

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

// --------------------------------------------------------------
//                    Reward Token Swap
// --------------------------------------------------------------

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

// --------------------------------------------------------------
//            Trade Route ICQ Test Helpers
// --------------------------------------------------------------

// Helper function to validate the address and denom from the query request data
func (s *KeeperTestSuite) validateAddressAndDenomInRequest(data []byte, expectedAddress, expectedDenom string) {
	actualAddress, actualDenom := s.ExtractAddressAndDenomFromBankPrefix(data)
	s.Require().Equal(expectedAddress, actualAddress, "query account address")
	s.Require().Equal(expectedDenom, actualDenom, "query denom")
}

// Helper function to validate the trade route query callback data
func (s *KeeperTestSuite) validateTradeRouteQueryCallback(actualCallbackDataBz []byte) {
	expectedCallbackData := types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	}

	var actualCallbackData types.TradeRouteCallback
	err := proto.Unmarshal(actualCallbackDataBz, &actualCallbackData)
	s.Require().NoError(err)
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callback data")
}

// --------------------------------------------------------------
//            Withdrawal Account - Reward Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the withdrawal address and the
// reward_denom_on_host since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceQueryTestCase() (route types.TradeRoute, expectedTimeout time.Duration) {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create and set the trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardDenomOnHostZone:   "ibc/reward_on_host",
		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
			Address:      StrideICAAddress, // must be a valid bech32, easiest to use stride prefix for validation
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts (the timeout is halfway through the epoch)
	epochDuration := time.Second * 30
	expectedTimeout = epochDuration / 2
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, epochDuration)

	return tradeRoute, expectedTimeout
}

// Tests a successful WithdrawalRewardBalanceQuery
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Successful() {
	route, timeoutDuration := s.SetupWithdrawalRewardBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when querying balance")

	// Validate fields from ICQ submission
	expectedRequestData := s.GetBankStoreKeyPrefix(StrideICAAddress, route.RewardDenomOnHostZone)

	query := s.ValidateQuerySubmission(
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_WithdrawalRewardBalance,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	s.validateAddressAndDenomInRequest(query.RequestData, route.HostAccount.Address, route.RewardDenomOnHostZone)
	s.validateTradeRouteQueryCallback(query.CallbackData)
}

// Tests a WithdrawalRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA account address to be invalid
	tradeRoute.HostAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid withdrawal account address")
}

// Tests a WithdrawalRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_MissingEpoch() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a WithdrawalRewardBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_FailedQuerySubmission() {
	tradeRoute, _ := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA connection id to be invalid
	tradeRoute.HostAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//             Trade Account - Reward Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
// reward_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeRewardBalanceQueryTestCase() (route types.TradeRoute, expectedTimeout time.Duration) {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create and set the trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardDenomOnTradeZone:  "ibc/reward_on_trade",
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
			Address:      StrideICAAddress, // must be a valid bech32, easiest to use stride prefix for validation
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts
	timeoutDuration := time.Second * 30
	s.CreateEpochForICATimeout(epochtypes.HOUR_EPOCH, timeoutDuration)

	return tradeRoute, timeoutDuration
}

// Tests a successful TradeRewardBalanceQuery
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Successful() {
	route, timeoutDuration := s.SetupTradeRewardBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when querying balance")

	// Validate fields from ICQ submission
	expectedRequestData := s.GetBankStoreKeyPrefix(StrideICAAddress, route.RewardDenomOnTradeZone)

	query := s.ValidateQuerySubmission(
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_TradeRewardBalance,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	s.validateAddressAndDenomInRequest(query.RequestData, route.TradeAccount.Address, route.RewardDenomOnTradeZone)
	s.validateTradeRouteQueryCallback(query.CallbackData)
}

// Tests a TradeRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute, _ := s.SetupTradeRewardBalanceQueryTestCase()

	// Change the trade ICA account address to be invalid
	tradeRoute.TradeAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_MissingEpoch() {
	tradeRoute, _ := s.SetupTradeRewardBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.HOUR_EPOCH)

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "hour: epoch not found")
}

// Tests a TradeRewardBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_FailedQuerySubmission() {
	tradeRoute, _ := s.SetupTradeRewardBalanceQueryTestCase()

	// Change the trade ICA connection id to be invalid
	tradeRoute.TradeAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//            Trade Account - Converted Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
// host_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeConvertedBalanceQueryTestCase() (route types.TradeRoute, expectedTimeout time.Duration) {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create and set the trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		HostDenomOnTradeZone:    "ibc/host_on_trade",
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
			Address:      StrideICAAddress, // must be a valid bech32, easiest to use stride prefix for validation
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create and set the epoch tracker for timeouts
	timeoutDuration := time.Second * 30
	s.CreateEpochForICATimeout(epochtypes.STRIDE_EPOCH, timeoutDuration)

	return tradeRoute, timeoutDuration
}

// Tests a successful TradeConvertedBalanceQuery
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Successful() {
	route, timeoutDuration := s.SetupTradeConvertedBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when querying balance")

	// Validate fields from ICQ submission
	expectedRequestData := s.GetBankStoreKeyPrefix(StrideICAAddress, route.HostDenomOnTradeZone)

	query := s.ValidateQuerySubmission(
		icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_TradeConvertedBalance,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	s.validateAddressAndDenomInRequest(query.RequestData, route.TradeAccount.Address, route.HostDenomOnTradeZone)
	s.validateTradeRouteQueryCallback(query.CallbackData)
}

// Tests a TradeConvertedBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA account address to be invalid
	tradeRoute.TradeAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeConvertedBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_MissingEpoch() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeConvertedBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_FailedQuerySubmission() {
	tradeRoute, _ := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA connection id to be invalid
	tradeRoute.TradeAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//                   Pool Price Query
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestPoolPriceQuery() {
	// Create a transfer channel so the connection exists for the query submission
	s.CreateTransferChannel(HostChainId)

	// Create an epoch tracker to dictate the query timeout
	timeoutDuration := time.Minute * 10
	s.CreateEpochForICATimeout(epochtypes.HOUR_EPOCH, timeoutDuration)

	// Define the trade route
	poolId := uint64(100)
	tradeRewardDenom := "ibc/reward-denom-on-trade"
	tradeHostDenom := "ibc/reward-denom-on-host"

	route := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		RewardDenomOnTradeZone:  tradeRewardDenom,
		HostDenomOnTradeZone:    tradeHostDenom,

		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
		},
		TradeConfig: types.TradeConfig{
			PoolId: poolId,
		},
	}

	expectedCallbackData := types.TradeRouteCallback{
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	}

	// Submit the pool price ICQ
	err := s.App.StakeibcKeeper.PoolPriceQuery(s.Ctx, route)
	s.Require().NoError(err, "no error expected when submitting pool price query")

	// Confirm the query request key is the same regardless of which order the denom's are specified
	expectedRequestData := icqtypes.FormatOsmosisMostRecentTWAPKey(poolId, tradeRewardDenom, tradeHostDenom)
	expectedRequestDataSwapped := icqtypes.FormatOsmosisMostRecentTWAPKey(poolId, tradeHostDenom, tradeRewardDenom)
	s.Require().Equal(expectedRequestData, expectedRequestDataSwapped, "osmosis twap denoms should be sorted")

	// Validate the fields of the query
	query := s.ValidateQuerySubmission(
		icqtypes.TWAP_STORE_QUERY_WITH_PROOF,
		expectedRequestData,
		keeper.ICQCallbackID_PoolPrice,
		timeoutDuration,
		icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	)

	// Validate the query callback data
	var actualCallbackData types.TradeRouteCallback
	err = proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err)
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callback data")

	// Remove the connection ID from the trade account and confirm the query submission fails
	invalidRoute := route
	invalidRoute.TradeAccount.ConnectionId = ""
	err = s.App.StakeibcKeeper.PoolPriceQuery(s.Ctx, invalidRoute)
	s.Require().ErrorContains(err, "invalid interchain query request")

	// Remove the epoch tracker so the function fails to get a timeout
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.HOUR_EPOCH)
	err = s.App.StakeibcKeeper.PoolPriceQuery(s.Ctx, route)
	s.Require().ErrorContains(err, "hour: epoch not found")
}
