package keeper_test

import (
	"fmt"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// ================================ 1: QueryValidatorExchangeRate =============================================

type QueryValidatorExchangeRateTestCase struct {
	msg                types.MsgUpdateValidatorSharesExchRate
	currentEpoch       uint64
	hostZone           types.HostZone
	strideEpochTracker types.EpochTracker
	dayEpochTracker    types.EpochTracker
}

func (s *KeeperTestSuite) SetupQueryValidatorExchangeRate() QueryValidatorExchangeRateTestCase {
	currentEpoch := uint64(1)
	valoperAddr := "cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz"

	// set up IBC
	s.CreateTransferChannel(HostChainId)

	hostZone := types.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		HostDenom:    Atom,
		IbcDenom:     IbcAtom,
		Bech32Prefix: Bech32Prefix,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// This will make the current time 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           10_000_000_000,                                               // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 1_000_000_000), // epoch ends in 1 second
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// This will make the current time 50% through the day
	dayEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           40_000_000_000,                                                // 40 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 20_000_000_000), // day ends in 20 second
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	return QueryValidatorExchangeRateTestCase{
		msg: types.MsgUpdateValidatorSharesExchRate{
			Creator: s.TestAccs[0].String(),
			ChainId: HostChainId,
			Valoper: valoperAddr,
		},
		currentEpoch:       currentEpoch,
		hostZone:           hostZone,
		strideEpochTracker: strideEpochTracker,
		dayEpochTracker:    dayEpochTracker,
	}
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_Successful() {
	tc := s.SetupQueryValidatorExchangeRate()

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().NoError(err, "no error expected")
	s.Require().NotNil(resp, "response should not be nil")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query should have been created")
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_BeforeBufferWindow() {
	tc := s.SetupQueryValidatorExchangeRate()

	// set the time to be 50% through the stride_epoch
	strideEpochTracker := tc.strideEpochTracker
	strideEpochTracker.NextEpochStartTime = uint64(s.Coordinator.CurrentTime.UnixNano() + int64(strideEpochTracker.Duration)/2) // 50% through the epoch
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "outside the buffer time during which ICQs are allowed")
	s.Require().Nil(resp, "response should be nil")
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_NoHostZone() {
	tc := s.SetupQueryValidatorExchangeRate()

	// remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, tc.hostZone.ChainId)

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "Host zone not found")
	s.Require().Nil(resp, "response should be nil")

	// submit a bad chain id
	tc.msg.ChainId = "NOT_GAIA"
	resp, err = s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "Host zone not found")
	s.Require().Nil(resp, "response should be nil")
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_ValoperDoesNotMatchBech32Prefix() {
	tc := s.SetupQueryValidatorExchangeRate()

	tc.msg.Valoper = "BADPREFIX_123"

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "validator operator address must match the host zone bech32 prefix")
	s.Require().Nil(resp, "response should be nil")
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_BadValoperAddress() {
	tc := s.SetupQueryValidatorExchangeRate()

	tc.msg.Valoper = "cosmos_BADADDRESS"

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "invalid validator operator address, could not decode")
	s.Require().Nil(resp, "response should be nil")
}

func (s *KeeperTestSuite) TestQueryValidatorExchangeRate_MissingConnectionId() {
	tc := s.SetupQueryValidatorExchangeRate()

	tc.hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	resp, err := s.App.StakeibcKeeper.QueryValidatorExchangeRate(s.Ctx, &tc.msg)
	s.Require().ErrorContains(err, "connection id cannot be empty")
	s.Require().Nil(resp, "response should be nil")
}

// ================================== 2: QueryDelegationsIcq ==========================================

type QueryDelegationsIcqTestCase struct {
	hostZone           types.HostZone
	valoperAddr        string
	strideEpochTracker types.EpochTracker
	dayEpochTracker    types.EpochTracker
}

func (s *KeeperTestSuite) SetupQueryDelegationsIcq() QueryDelegationsIcqTestCase {
	currentEpoch := uint64(1)
	valoperAddr := "cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz"

	// set up IBC
	s.CreateTransferChannel(HostChainId)

	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	hostZone := types.HostZone{
		ChainId:           HostChainId,
		ConnectionId:      ibctesting.FirstConnectionID,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		Bech32Prefix:      Bech32Prefix,
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// This will make the current time 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           10_000_000_000,                                               // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 1_000_000_000), // epoch ends in 1 second
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// This will make the current time 50% through the day
	dayEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           40_000_000_000,                                                // 40 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 20_000_000_000), // day ends in 20 second
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	return QueryDelegationsIcqTestCase{
		hostZone:           hostZone,
		valoperAddr:        valoperAddr,
		strideEpochTracker: strideEpochTracker,
		dayEpochTracker:    dayEpochTracker,
	}
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_Successful() {
	tc := s.SetupQueryDelegationsIcq()

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query should have been created")

	// querying twice with the same query should only create one query
	err = s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "querying twice with the same query should only create one query")

	// querying with a different query should create a second query
	tc.valoperAddr = "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne"
	err = s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 2, "querying with a different query should create a second query")
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_BeforeBufferWindow() {
	tc := s.SetupQueryDelegationsIcq()

	// set the time to be 50% through the stride_epoch
	strideEpochTracker := tc.strideEpochTracker
	strideEpochTracker.NextEpochStartTime = uint64(s.Coordinator.CurrentTime.UnixNano() + int64(strideEpochTracker.Duration)/2) // 50% through the epoch
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().ErrorContains(err, "outside the buffer time during which ICQs are allowed")
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_MissingDelegationAddress() {
	tc := s.SetupQueryDelegationsIcq()

	tc.hostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().ErrorContains(err, "missing a delegation address")
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_MissingConnectionId() {
	tc := s.SetupQueryDelegationsIcq()

	tc.hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, tc.hostZone, tc.valoperAddr)
	s.Require().ErrorContains(err, "connection id cannot be empty")
}
