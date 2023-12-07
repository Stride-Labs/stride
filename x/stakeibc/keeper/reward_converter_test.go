package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

type ICQCallbackArgs struct {
	Query        icqtypes.Query
	CallbackArgs []byte
}

type BalanceQueryCallbackTestCase struct {
	TradeRoute types.TradeRoute
	Response   ICQCallbackArgs
	ChannelID  string
	PortID     string
}

// --------------------------------------------------------------
//                   Transfer Host to Trade
// --------------------------------------------------------------

// Tests TransferRewardTokensHostToTrade and BuildHostToTradeTransferMsg
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

	// Delete the epoch tracker and call each function, confirming they both fail
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err = s.App.StakeibcKeeper.BuildHostToTradeTransferMsg(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
	err = s.App.StakeibcKeeper.TransferRewardTokensHostToTrade(s.Ctx, transferAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
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
}

func (s *KeeperTestSuite) TestSwapRewardTokens() {
	// Create an ICA channel for the transfer submission
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_CONVERTER_TRADE)
	channelId, portId := s.CreateICAChannel(owner)

	minSwapAmount := sdkmath.NewInt(10)
	rewardAmount := sdkmath.NewInt(100)

	route := types.TradeRoute{
		RewardDenomOnTradeZone: "ibc/reward_on_trade",
		HostDenomOnTradeZone:   "ibc/host_on_trade",

		TradeAccount: types.ICAAccount{
			Address:      "trade_address",
			ConnectionId: ibctesting.FirstConnectionID,
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
	s.CreateStrideEpochForICATimeout(time.Minute) // arbitrary timeout time

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
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)
	err = s.App.StakeibcKeeper.SwapRewardTokens(s.Ctx, rewardAmount, route)
	s.Require().ErrorContains(err, "epoch not found")
}

// --------------------------------------------------------------
//            Withdrawal Account - Reward Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the withdrawal address and the
//
//	reward_denom_on_host since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceQueryTestCase() types.TradeRoute {
	// Create the connection between Stride and HostChain with the withdrawal account initialized
	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, types.ICAAccountType_WITHDRAWAL.String())
	withdrawalChannelId, withdrawalPortId := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]
	withdrawalConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, withdrawalPortId, withdrawalChannelId)

	hostICA := types.ICAAccount{
		ChainId:      HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: withdrawalConnectionId,
		Address:      withdrawalAddress,
	}

	// Must initialize these or they will serialize differently from the default values
	tradeConfig := types.TradeConfig{
		SwapPrice:              sdk.OneDec(),
		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.05"),
		MinSwapAmount:          sdk.ZeroInt(),
		MaxSwapAmount:          sdk.NewIntFromUint64(uint64(1_000_000)),
	}

	// Create and set the trade route for testing
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone: "ibc/reward_on_host",
		HostAccount:           hostICA,
		TradeConfig:           tradeConfig,
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

	return tradeRoute
}

// Tests a successful WithdrawalRewardBalanceQuery
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Successful() {
	tradeRoute := s.SetupWithdrawalRewardBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tradeRoute.HostAccount.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tradeRoute.HostAccount.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_WithdrawalRewardBalance, query.CallbackId, "query callback id")

	var actualCallbackData types.TradeRoute
	err = proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	// Callback data should just be the trade route itself
	s.Require().Equal(tradeRoute, actualCallbackData, "query callabck data")

	// Confirm query request info
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(tradeRoute.HostAccount.Address, actualAddress.String(), "query account address")
	s.Require().Equal(tradeRoute.RewardDenomOnHostZone, actualDenom, "query denom")
}

// Tests a WithdrawalRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA account address to be invalid
	tradeRoute.HostAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid withdrawal account address")
}

// Tests a WithdrawalRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_MissingEpoch() {
	tradeRoute := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a WithdrawalRewardBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_FailedQuerySubmission() {
	tradeRoute := s.SetupWithdrawalRewardBalanceQueryTestCase()

	// Change the withdrawal ICA connection id to be invalid
	tradeRoute.HostAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//             Trade Account - Reward Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
//
//	reward_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeRewardBalanceQueryTestCase() types.TradeRoute {
	// Create the connection between Stride and OsmoChain with the trade account initialized
	tradeAccountOwner := fmt.Sprintf("%s.%s", OsmoChainId, types.ICAAccountType_CONVERTER_TRADE.String())
	tradeChannelId, tradePortId := s.CreateICAChannel(tradeAccountOwner)
	tradeAddress := s.IcaAddresses[tradeAccountOwner]
	tradeConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, tradePortId, tradeChannelId)

	tradeICA := types.ICAAccount{
		ChainId:      OsmoChainId,
		Type:         types.ICAAccountType_CONVERTER_TRADE,
		ConnectionId: tradeConnectionId,
		Address:      tradeAddress,
	}

	// Must initialize these or they will serialize differently from the default values
	tradeConfig := types.TradeConfig{
		SwapPrice:              sdk.OneDec(),
		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.05"),
		MinSwapAmount:          sdk.ZeroInt(),
		MaxSwapAmount:          sdk.NewIntFromUint64(uint64(1_000_000)),
	}

	// Create and set the trade route for testing
	tradeRoute := types.TradeRoute{
		RewardDenomOnTradeZone: "ibc/reward_on_trade",
		TradeAccount:           tradeICA,
		TradeConfig:            tradeConfig,
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

	return tradeRoute
}

// Tests a successful TradeRewardBalanceQuery
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Successful() {
	tradeRoute := s.SetupTradeRewardBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tradeRoute.TradeAccount.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tradeRoute.TradeAccount.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_TradeRewardBalance, query.CallbackId, "query callback id")

	var actualCallbackData types.TradeRoute
	err = proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	// Callback data should just be the trade route itself
	s.Require().Equal(tradeRoute, actualCallbackData, "query callabck data")

	// Confirm query request info
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(tradeRoute.TradeAccount.Address, actualAddress.String(), "query account address")
	s.Require().Equal(tradeRoute.RewardDenomOnTradeZone, actualDenom, "query denom")
}

// Tests a TradeRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute := s.SetupTradeRewardBalanceQueryTestCase()

	// Change the trade ICA account address to be invalid
	tradeRoute.TradeAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_MissingEpoch() {
	tradeRoute := s.SetupTradeRewardBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeRewardBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_FailedQuerySubmission() {
	tradeRoute := s.SetupTradeRewardBalanceQueryTestCase()

	// Change the trade ICA connection id to be invalid
	tradeRoute.TradeAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//            Trade Account - Converted Balance Query
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
//
//	host_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeConvertedBalanceQueryTestCase() types.TradeRoute {
	// Create the connection between Stride and OsmoChain with the trade account initialized
	tradeAccountOwner := fmt.Sprintf("%s.%s", OsmoChainId, types.ICAAccountType_CONVERTER_TRADE.String())
	tradeChannelId, tradePortId := s.CreateICAChannel(tradeAccountOwner)
	tradeAddress := s.IcaAddresses[tradeAccountOwner]
	tradeConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, tradePortId, tradeChannelId)

	tradeICA := types.ICAAccount{
		ChainId:      OsmoChainId,
		Type:         types.ICAAccountType_CONVERTER_TRADE,
		ConnectionId: tradeConnectionId,
		Address:      tradeAddress,
	}

	// Must initialize these or they will serialize differently from the default values
	tradeConfig := types.TradeConfig{
		SwapPrice:              sdk.OneDec(),
		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr("0.05"),
		MinSwapAmount:          sdk.ZeroInt(),
		MaxSwapAmount:          sdk.NewIntFromUint64(uint64(1_000_000)),
	}

	// Create and set the trade route for testing
	tradeRoute := types.TradeRoute{
		HostDenomOnTradeZone: "ibc/host_on_trade",
		TradeAccount:         tradeICA,
		TradeConfig:          tradeConfig,
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

	return tradeRoute
}

// Tests a successful TradeConvertedBalanceQuery
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Successful() {
	tradeRoute := s.SetupTradeConvertedBalanceQueryTestCase()

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tradeRoute.TradeAccount.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tradeRoute.TradeAccount.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_TradeConvertedBalance, query.CallbackId, "query callback id")

	var actualCallbackData types.TradeRoute
	err = proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	// Callback data should just be the trade route itself
	s.Require().Equal(tradeRoute, actualCallbackData, "query callabck data")

	// Confirm query request info
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(tradeRoute.TradeAccount.Address, actualAddress.String(), "query account address")
	s.Require().Equal(tradeRoute.HostDenomOnTradeZone, actualDenom, "query denom")
}

// Tests a TradeConvertedBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_InvalidAccountAddress() {
	tradeRoute := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA account address to be invalid
	tradeRoute.TradeAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeConvertedBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_MissingEpoch() {
	tradeRoute := s.SetupTradeConvertedBalanceQueryTestCase()

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeConvertedBalanceQuery that fails to submit the query due to bad connection
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_FailedQuerySubmission() {
	tradeRoute := s.SetupTradeConvertedBalanceQueryTestCase()

	// Change the trade ICA connection id to be invalid
	tradeRoute.TradeAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// --------------------------------------------------------------
//                   Pool Price Query
// --------------------------------------------------------------
