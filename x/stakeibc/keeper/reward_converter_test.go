package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"

	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// --------------------------------------------------------------
// Tests for WithdrawalRewardBalanceQuery ICQ performing function
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the withdrawal address and the
//  reward_denom_on_host since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupWithdrawalRewardBalanceQueryTestCase() types.TradeRoute {
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
		HostAccount: 			hostICA,
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
// Tests for TradeRewardBalanceQuery ICQ performing function
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
//  reward_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeRewardBalanceQueryTestCase() types.TradeRoute {
	// Create the connection between Stride and OsmoChain with the trade account initialized
	tradeAccountOwner := fmt.Sprintf("%s.%s", OsmoChainId, types.ICAAccountType_CONVERTER_TRADE.String())
	tradeChannelId, tradePortId := s.CreateICAChannel(tradeAccountOwner)
	tradeAddress := s.IcaAddresses[tradeAccountOwner]
	tradeConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, tradePortId, tradeChannelId)

	tradeICA := types.ICAAccount{
		ChainId: OsmoChainId,
		Type: types.ICAAccountType_CONVERTER_TRADE,
		ConnectionId: tradeConnectionId,
		Address: tradeAddress,
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
		RewardDenomOnTradeZone: "ibc/reward_on_trade",
		TradeAccount: 			tradeICA,
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
// Tests for TradeConvertedBalanceQuery ICQ performing function
// --------------------------------------------------------------

// Create the traderoute for these tests, only need the trade address and the
//  host_denom_on_trade since this will be what is used in the query, no other setup
func (s *KeeperTestSuite) SetupTradeConvertedBalanceQueryTestCase() types.TradeRoute {
	// Create the connection between Stride and OsmoChain with the trade account initialized
	tradeAccountOwner := fmt.Sprintf("%s.%s", OsmoChainId, types.ICAAccountType_CONVERTER_TRADE.String())
	tradeChannelId, tradePortId := s.CreateICAChannel(tradeAccountOwner)
	tradeAddress := s.IcaAddresses[tradeAccountOwner]
	tradeConnectionId, _, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.
		GetChannelConnection(s.Ctx, tradePortId, tradeChannelId)

	tradeICA := types.ICAAccount{
		ChainId: OsmoChainId,
		Type: types.ICAAccountType_CONVERTER_TRADE,
		ConnectionId: tradeConnectionId,
		Address: tradeAddress,
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
		HostDenomOnTradeZone: 	"ibc/host_on_trade",
		TradeAccount: 			tradeICA,
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







// The functions which build and send off the ICAs for token transfers and swap commands
// all use the SubmitTxWithoutCallback, so no callback data is stored to examine
// We only need to test that an IBC message was indeed sent

// Since each of the ICA functions is called from the icqcallback functions
// We can test these callback functions directly and implictly test the ICA building funcs

// Useful structures used by the icqcallback tests
type ICQCallbackArgs struct {
	Query        icqtypes.Query
	CallbackArgs []byte
}

type BalanceQueryCallbackTestCase struct {
	TradeRoute	types.TradeRoute
	Response	ICQCallbackArgs
	ChannelID	string
	PortID		string
}

// The callback tests are in their own appropriately named files, similar to other callback tests
// WithdrawalRewardBalanceCallback will trigger TransferRewardTokensHostToTrade
// TradeRewardBalanceCallback will trigger SwapRewardTokens
// TradeConvertedBalanceCallback will trigger TransferConvertedTokensTradeToHost
