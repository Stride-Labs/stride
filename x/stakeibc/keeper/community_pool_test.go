package keeper_test

import (
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type QueryCommunityPoolBalanceTestCase struct {
	hostZone        types.HostZone
	timeoutDuration time.Duration
	expectedTimeout uint64
}

func (s *KeeperTestSuite) SetupQueryCommunityPoolBalance(icaAccountType types.ICAAccountType) QueryCommunityPoolBalanceTestCase {
	// We need to register the transfer channel to initialize the light client state
	s.CreateTransferChannel(HostChainId)

	// Create host zone
	// We must use valid addresses for each ICA since they're serialized for the query request
	depositAddress := s.TestAccs[0]
	returnAddress := s.TestAccs[1]
	hostZone := types.HostZone{
		ChainId:                        HostChainId,
		ConnectionId:                   ibctesting.FirstConnectionID,
		CommunityPoolDepositIcaAddress: depositAddress.String(),
		CommunityPoolReturnIcaAddress:  returnAddress.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create epoch tracker for timeout
	timeoutDuration := time.Second * 30
	epochEndTime := uint64(s.Ctx.BlockTime().Add(timeoutDuration).UnixNano())
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		NextEpochStartTime: epochEndTime,
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	return QueryCommunityPoolBalanceTestCase{
		hostZone:        hostZone,
		timeoutDuration: timeoutDuration,
		expectedTimeout: epochEndTime,
	}
}

// Helper function to verify the query that was submitted from the community pool balance query
func (s *KeeperTestSuite) checkCommunityPoolQuerySubmission(
	tc QueryCommunityPoolBalanceTestCase,
	icaAccountType types.ICAAccountType,
	expectedIcaAddress string,
) {
	// Check that one query was submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "there should have been 1 query submitted")
	query := queries[0]

	// Confirm query contents
	s.Require().Equal(tc.hostZone.ChainId, query.ChainId, "query chain ID")
	s.Require().Equal(tc.hostZone.ConnectionId, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE, query.TimeoutPolicy, "query timeout policy")

	// Confirm query timeout info
	s.Require().Equal(tc.timeoutDuration, query.TimeoutDuration, "query callback id")
	s.Require().Equal(tc.expectedTimeout, query.TimeoutTimestamp, "query callback id")

	// Confirm callback data
	s.Require().Equal(types.ModuleName, query.CallbackModule, "query callback module")
	s.Require().Equal(keeper.ICQCallbackID_CommunityPoolBalance, query.CallbackId, "query callback id")

	var actualCallbackData types.CommunityPoolBalanceQueryCallback
	err := proto.Unmarshal(query.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	expectedCallbackData := types.CommunityPoolBalanceQueryCallback{
		IcaType: icaAccountType,
		Denom:   Atom,
	}
	s.Require().Equal(expectedCallbackData, actualCallbackData, "query callabck data")

	// Confirm query request info
	requestData := query.RequestData[1:] // Remove BalancePrefix byte
	actualAddress, actualDenom, err := banktypes.AddressAndDenomFromBalancesStore(requestData)
	s.Require().NoError(err, "no error expected when retrieving address and denom from store key")
	s.Require().Equal(expectedIcaAddress, actualAddress.String(), "query account address")
	s.Require().Equal(Atom, actualDenom, "query denom")
}

// Tests a community pool balance query to the deposit ICA account
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Successful_Deposit() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().NoError(err, "no error expected when querying pool balance")

	s.checkCommunityPoolQuerySubmission(tc, icaAccountType, tc.hostZone.CommunityPoolDepositIcaAddress)
}

// Tests a community pool balance query to the return ICA account
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Successful_Return() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_RETURN
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().NoError(err, "no error expected when querying pool balance")

	s.checkCommunityPoolQuerySubmission(tc, icaAccountType, tc.hostZone.CommunityPoolReturnIcaAddress)
}

// Tests a community pool balance query that fails due to an invalid account type
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_InvalidAccountType() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	invalidAccountType := types.ICAAccountType_DELEGATION
	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, tc.hostZone, invalidAccountType, Atom)
	s.Require().ErrorContains(err, "icaType must be either deposit or return!")
}

// Tests a community pool balance query that fails due to an invalid account address
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_InvalidAccountAddress() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Change the host zone account address to be invalid
	invalidHostZone := tc.hostZone
	invalidHostZone.CommunityPoolDepositIcaAddress = "invalid_address"

	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, invalidHostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "invalid COMMUNITY_POOL_DEPOSIT address, could not decode (invalid_address)")
}

// Tests a community pool balance query that fails due to a missing epoch tracker
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_Failure_MissingEpoch() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Remove the stride epoch so the test fails
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, tc.hostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "stride_epoch: epoch not found")
}

// Tests a community pool balance query that fails to submit the query
func (s *KeeperTestSuite) TestQueryCommunityPoolBalance_FailedQuerySubmission() {
	icaAccountType := types.ICAAccountType_COMMUNITY_POOL_DEPOSIT
	tc := s.SetupQueryCommunityPoolBalance(icaAccountType)

	// Set an invalid connection ID for the host zone so that the query submission fails
	invalidHostZone := tc.hostZone
	invalidHostZone.ConnectionId = "invalid_connection"

	err := s.App.StakeibcKeeper.QueryCommunityPoolBalance(s.Ctx, invalidHostZone, icaAccountType, Atom)
	s.Require().ErrorContains(err, "Error submitting query for pool ica balance")
}
