package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// ---------------------------------------
// Query Balances for denoms on TradeRoute
// ---------------------------------------

type QueryRouteICABalanceTestCase struct {
	icqCallbackId     string
	tradeRoute        types.TradeRoute
	testICA			  types.ICAAccount
	testDenom         string
	timeoutDuration   time.Duration
	expectedTimeout   uint64
}

func (s *KeeperTestSuite) SetupQueryRouteICABalance(icqCallbackId string) QueryRouteICABalanceTestCase {
	// We need to register the transfer channel to initialize the light client state
	switch icqCallbackId {
    case keeper.ICQCallbackID_WithdrawalRewardBalance:
		s.CreateTransferChannel(HostChainId)
	case keeper.ICQCallbackID_TradeRewardBalance, keeper.ICQCallbackID_TradeConvertedBalance:
		s.CreateTransferChannel(OsmoChainId)

    }	

	// Create relevant TradeRoute fields
	// We must use valid addresses for each ICA since they're serialized for the query request
	withdrawalAddress := s.TestAccs[0].String()
	tradeAddress := s.TestAccs[1].String()

	hostICA := types.ICAAccount{
		ChainId: HostChainId,
		Type: types.ICAAccountType_WITHDRAWAL,
		ConnectionId: ibctesting.FirstConnectionID,
		Address: withdrawalAddress,
	}

	rewardICA := types.ICAAccount{}

	tradeICA := types.ICAAccount{
		ChainId: OsmoChainId,
		Type: types.ICAAccountType_TRADE,
		ConnectionId: ibctesting.FirstConnectionID,
		Address: tradeAddress,		
	}

	hostRewardHop := types.TradeHop{
		FromAccount: hostICA,
		ToAccount: rewardICA,
	}

	rewardTradeHop := types.TradeHop{
		FromAccount: rewardICA,
		ToAccount: tradeICA,
	}

	tradeHostHop := types.TradeHop{
		FromAccount: tradeICA,
		ToAccount: hostICA,
	}

	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone: "ibc/reward_on_host",
		RewardDenomOnTradeZone: "ibc/reward_on_trade",
		TargetDenomOnTradeZone: "ibc/host_on_trade",
		TargetDenomOnHostZone: "host_denom", // needed to save and get in keeper
		HostToRewardHop: hostRewardHop,
		RewardToTradeHop: rewardTradeHop,
		TradeToHostHop: tradeHostHop,
		MinSwapAmount: sdk.ZeroInt(), // if we don't initialize these they marshal weird
		MaxSwapAmount: sdk.ZeroInt(), // if we don't initialize these they marshal weird
	}

	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Create epoch tracker for timeout
	timeoutDuration := time.Second * 30
	epochEndTime := uint64(s.Ctx.BlockTime().Add(timeoutDuration).UnixNano())
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: epochEndTime,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	testDenom := ""
	testICA := types.ICAAccount{}
	
	switch icqCallbackId {
    case keeper.ICQCallbackID_WithdrawalRewardBalance:
		testDenom = tradeRoute.RewardDenomOnHostZone
        testICA = hostICA
	case keeper.ICQCallbackID_TradeRewardBalance:
		testDenom = tradeRoute.RewardDenomOnTradeZone
		testICA = tradeICA
	case keeper.ICQCallbackID_TradeConvertedBalance:
		testDenom = tradeRoute.TargetDenomOnTradeZone		
		testICA = tradeICA
    }


	return QueryRouteICABalanceTestCase{
		icqCallbackId:   icqCallbackId,
		tradeRoute:      tradeRoute,
		testICA:         testICA,
		testDenom:       testDenom,
		timeoutDuration: timeoutDuration,
		expectedTimeout: epochEndTime,
	}
}

// General helper function to verify submitted query for any of the three main query types
// Use the callbackId to determine which because all 3 queries have the same structure
func (s *KeeperTestSuite) checkRouteICABalanceQuerySubmission(
	tc QueryRouteICABalanceTestCase,
) {
	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tc.testICA.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tc.testICA.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm query timeout info
	s.Require().Equal(tc.timeoutDuration, query.TimeoutDuration, "query callback id")
	s.Require().Equal(tc.expectedTimeout, query.TimeoutTimestamp, "query callback id")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(tc.icqCallbackId, query.CallbackId, "query callback id")

	var actualCallbackData types.TradeRoute
	err := proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	expectedCallbackData := tc.tradeRoute //the callback data should just be the trade route itself
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callabck data")

	// Confirm query request info
	expectedIcaAddress := tc.testICA.Address
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(expectedIcaAddress, actualAddress.String(), "query account address")
	s.Require().Equal(tc.testDenom, actualDenom, "query denom")
}


// Use the general helper to test each of
// - WithdrawalRewardBalanceQuery
// - TradeRewardBalanceQuery
// - TradeConvertedBalanceQuery
// because they all have the same structure except which ICA and which denom is used


// Tests a WithdrawalRewardBalanceQuery
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Successful() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_WithdrawalRewardBalance)

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	s.checkRouteICABalanceQuerySubmission(tc)
}

// Tests a TradeRewardBalanceQuery
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Successful() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeRewardBalance)

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	s.checkRouteICABalanceQuerySubmission(tc)
}

// Tests a TradeConvertedBalanceQuery
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Successful() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeConvertedBalance)

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().NoError(err, "no error expected when querying balance")

	s.checkRouteICABalanceQuerySubmission(tc)
}



// Tests a WithdrawalRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_WithdrawalRewardBalance)

	// Change the withdrawal ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.HostToRewardHop.FromAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid withdrawal account address")
}

// Tests a TradeRewardBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_InvalidAccountAddress() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeRewardBalance)

	// Change the trade ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.RewardToTradeHop.ToAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}

// Tests a TradeConvertedBalanceQuery that fails due to an invalid account address
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_InvalidAccountAddress() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeConvertedBalance)

	// Change the trade ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.RewardToTradeHop.ToAccount.Address = "invalid_address"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid trade account address")
}





// Tests a WithdrawalRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_Failure_MissingEpoch() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_WithdrawalRewardBalance)

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeRewardBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_MissingEpoch() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeRewardBalance)

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a TradeConvertedBalanceQuery that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_MissingEpoch() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeConvertedBalance)

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, tc.tradeRoute)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}






// Tests a WithdrawalRewardBalanceQuery that fails to submit the query
func (s *KeeperTestSuite) TestWithdrawalRewardBalanceQuery_FailedQuerySubmission() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_WithdrawalRewardBalance)

	// Change the withdrawal ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.HostToRewardHop.FromAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.WithdrawalRewardBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// Tests a TradeRewardBalanceQuery that fails to submit the query
func (s *KeeperTestSuite) TestTradeRewardBalanceQuery_Failure_FailedQuerySubmission() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeRewardBalance)

	// Change the trade ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.RewardToTradeHop.ToAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeRewardBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}

// Tests a TradeConvertedBalanceQuery that fails to submit the query
func (s *KeeperTestSuite) TestTradeConvertedBalanceQuery_Failure_FailedQuerySubmission() {
	tc := s.SetupQueryRouteICABalance(keeper.ICQCallbackID_TradeConvertedBalance)

	// Change the trade ICA account address to be invalid
	invalidTradeRoute := tc.tradeRoute
	invalidTradeRoute.RewardToTradeHop.ToAccount.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.TradeConvertedBalanceQuery(s.Ctx, invalidTradeRoute)
	s.Require().ErrorContains(err, "invalid connection-id (invalid_connection)")
}
