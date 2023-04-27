package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck
	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// ================================ 1: QueryValidatorExchangeRate =============================================

type QueryValidatorExchangeRateTestCase struct {
	msg      types.MsgUpdateValidatorSharesExchRate
	hostZone types.HostZone
}

func (s *KeeperTestSuite) SetupQueryValidatorExchangeRate() QueryValidatorExchangeRateTestCase {
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

	return QueryValidatorExchangeRateTestCase{
		msg: types.MsgUpdateValidatorSharesExchRate{
			Creator: s.TestAccs[0].String(),
			ChainId: HostChainId,
			Valoper: valoperAddr,
		},
		hostZone: hostZone,
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
	s.Require().ErrorContains(err, "connection-id cannot be empty")
	s.Require().Nil(resp, "response should be nil")
}

// ================================== 2: QueryDelegationsIcq ==========================================

func (s *KeeperTestSuite) SetupQueryDelegationsIcq() (types.HostZone, types.Validator) {
	// set up IBC
	s.CreateTransferChannel(HostChainId)

	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		IbcDenom:             IbcAtom,
		Bech32Prefix:         Bech32Prefix,
		DelegationIcaAddress: delegationAddress,
	}
	validator := types.Validator{Address: ValAddress, BalancedDelegation: sdkmath.NewInt(100)}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return hostZone, validator
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_Successful() {
	hostZone, validator := s.SetupQueryDelegationsIcq()

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, hostZone, validator, uint64(1))
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query should have been created")

	// confirm callback data from query
	var callbackData types.DelegatorSharesQueryCallback
	err = proto.Unmarshal(queries[0].CallbackData, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")
	s.Require().Equal(validator.BalancedDelegation, callbackData.InitialValidatorDelegation, "query callback data delegation")

	// querying twice with the same query should only create one query
	err = s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, hostZone, validator, uint64(1))
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "querying twice with the same query should only create one query")

	// querying with a different query should create a second query
	differentValidator := types.Validator{
		Address: "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne",
	}
	err = s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, hostZone, differentValidator, uint64(1))
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 2, "querying with a different query should create a second query")
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_MissingDelegationAddress() {
	hostZone, validator := s.SetupQueryDelegationsIcq()

	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, hostZone, validator, uint64(1))
	s.Require().ErrorContains(err, "no delegation address found for")
}

func (s *KeeperTestSuite) TestQueryDelegationsIcq_MissingConnectionId() {
	hostZone, validator := s.SetupQueryDelegationsIcq()

	hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := s.App.StakeibcKeeper.QueryDelegationsIcq(s.Ctx, hostZone, validator, uint64(1))
	s.Require().ErrorContains(err, "connection-id cannot be empty")
}
