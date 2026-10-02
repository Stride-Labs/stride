package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
	_ "github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// -----------------------------------------------------------
//	           QueryValidatorSharesToTokensRate
// -----------------------------------------------------------

type QueryValidatorSharesToTokensRateTestCase struct {
	hostZone types.HostZone
}

func (s *KeeperTestSuite) SetupQueryValidatorSharesToTokensRate() QueryValidatorSharesToTokensRateTestCase {
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

	return QueryValidatorSharesToTokensRateTestCase{
		hostZone: hostZone,
	}
}

func (s *KeeperTestSuite) TestQueryValidatorSharesToTokensRate_Successful() {
	s.SetupQueryValidatorSharesToTokensRate()

	err := s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, HostChainId, ValAddress)
	s.Require().NoError(err, "no error expected when querying validator sharesToTokens rate")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query should have been created")
}

func (s *KeeperTestSuite) TestQueryValidatorSharesToTokensRate_NoHostZone() {
	s.SetupQueryValidatorSharesToTokensRate()

	// remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, HostChainId, ValAddress)
	s.Require().ErrorContains(err, "Host zone not found")

	// submit a bad chain id
	err = s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, "NOT_GAIA", ValAddress)
	s.Require().ErrorContains(err, "Host zone not found")
}

func (s *KeeperTestSuite) TestQueryValidatorSharesToTokensRate_InvalidValidator() {
	s.SetupQueryValidatorSharesToTokensRate()

	// Pass a validator with an invalid prefix - it should fail
	err := s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, HostChainId, "BADPREFIX_123")
	s.Require().ErrorContains(err, "validator operator address must match the host zone bech32 prefix")

	// Pass a validator with a valid prefix but an invalid address - it should fail
	err = s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, HostChainId, "cosmos_BADADDRESS")
	s.Require().ErrorContains(err, "invalid validator operator address, could not decode")
}

func (s *KeeperTestSuite) TestQueryValidatorSharesToTokensRate_MissingConnectionId() {
	tc := s.SetupQueryValidatorSharesToTokensRate()

	tc.hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	err := s.App.StakeibcKeeper.QueryValidatorSharesToTokensRate(s.Ctx, HostChainId, ValAddress)
	s.Require().ErrorContains(err, "connection-id cannot be empty")
}

// -----------------------------------------------------------
//	                 SubmitDelegationICQ
// -----------------------------------------------------------

func (s *KeeperTestSuite) SetupSubmitDelegationICQ() (types.HostZone, types.Validator) {
	// set up IBC
	s.CreateTransferChannel(HostChainId)

	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	queriedValidator := types.Validator{
		Address:              ValAddress,
		Delegation:           sdkmath.NewInt(100),
		SlashQueryInProgress: false,
	}
	otherValidator := types.Validator{
		Address:              "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne",
		Delegation:           sdkmath.NewInt(100),
		SlashQueryInProgress: false,
	}
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		IbcDenom:             IbcAtom,
		Bech32Prefix:         Bech32Prefix,
		DelegationIcaAddress: delegationAddress,
		Validators:           []*types.Validator{&queriedValidator, &otherValidator},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return hostZone, queriedValidator
}

func (s *KeeperTestSuite) TestSubmitDelegationICQ_Successful() {
	hostZone, validator := s.SetupSubmitDelegationICQ()

	err := s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, ValAddress)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query should have been created")

	// confirm callback data from query
	var callbackData types.DelegatorSharesQueryCallback
	err = proto.Unmarshal(queries[0].CallbackData, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")
	s.Require().Equal(validator.Delegation, callbackData.InitialValidatorDelegation, "query callback data delegation")

	// querying twice with the same query should only create one query
	err = s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, ValAddress)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "querying twice with the same query should only create one query")

	// querying with a different query should create a second query
	differentValidator := hostZone.Validators[1].Address
	err = s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, differentValidator)
	s.Require().NoError(err, "no error expected")

	// check a query was created (a simple test; details about queries are covered in makeRequest's test)
	queries = s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 2, "querying with a different query should create a second query")
}

func (s *KeeperTestSuite) TestSubmitCalibrationICQ_SnapshotsDelegation() {
	hostZone, validator := s.SetupSubmitDelegationICQ()
	before := s.MustGetHostZone(HostChainId)
	err := s.App.StakeibcKeeper.SubmitCalibrationICQ(s.Ctx, hostZone, validator.Address)
	s.Require().NoError(err)
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1)
	var callbackData types.DelegatorSharesQueryCallback
	s.Require().NoError(proto.Unmarshal(queries[0].CallbackData, &callbackData))
	s.Require().Equal(validator.Delegation, callbackData.InitialValidatorDelegation)
	s.Require().Equal(validator.Address, callbackData.ValidatorAddress, "an empty response has no Delegation to read the validator from")
	s.Require().True(queries[0].InvokeCallbackOnEmptyResponse, "calibration opts into the callback for an empty response")
	s.Require().Equal(before, s.MustGetHostZone(HostChainId), "snapshot creation does not change validator accounting")
}

// Only calibration opts in: an empty slash-query response keeps being dropped
func (s *KeeperTestSuite) TestSubmitDelegationICQ_DoesNotOptIntoEmptyResponse() {
	hostZone, validator := s.SetupSubmitDelegationICQ()
	s.Require().NoError(s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, validator.Address))
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1)
	s.Require().False(queries[0].InvokeCallbackOnEmptyResponse)
}

func (s *KeeperTestSuite) TestSubmitCalibrationICQ_PurgedLegacyIDRemainsAbsent() {
	hostZone, validator := s.SetupSubmitDelegationICQ()
	_, validatorAddress, err := bech32.DecodeAndConvert(validator.Address)
	s.Require().NoError(err)
	_, delegatorAddress, err := bech32.DecodeAndConvert(hostZone.DelegationIcaAddress)
	s.Require().NoError(err)
	legacyQuery := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     stakingtypes.GetDelegationKey(delegatorAddress, validatorAddress),
		CallbackModule:  types.ModuleName,
		CallbackId:      keeper.ICQCallbackID_Calibrate,
		TimeoutDuration: time.Hour,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	legacyID := s.App.InterchainqueryKeeper.GetQueryId(s.Ctx, legacyQuery, false)
	s.Require().NoError(s.App.InterchainqueryKeeper.SubmitICQRequest(s.Ctx, legacyQuery, false))
	_, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, legacyID)
	s.Require().True(found, "legacy submission uses the real deterministic request hash")
	v35.PurgeCalibrationQueries(s.Ctx, s.App.InterchainqueryKeeper)
	_, found = s.App.InterchainqueryKeeper.GetQuery(s.Ctx, legacyID)
	s.Require().False(found)

	// A new admin query for the same key must not revive the purged response's query ID.
	s.Require().NoError(s.App.StakeibcKeeper.SubmitCalibrationICQ(s.Ctx, hostZone, validator.Address))
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1)
	freshQuery := queries[0]
	before := s.MustGetHostZone(HostChainId)
	queryResponse := s.CreateDelegatorSharesQueryResponse(validator.Address, sdkmath.LegacyNewDec(200))
	_, err = icqkeeper.NewMsgServerImpl(s.App.InterchainqueryKeeper).SubmitQueryResponse(s.Ctx,
		&icqtypes.MsgSubmitQueryResponse{QueryId: legacyID, Result: queryResponse})
	s.Require().NoError(err, "a late legacy response bypasses proof and callback processing as a successful no-op")
	s.Require().NotEqual(legacyID, freshQuery.Id)
	s.Require().Equal(before, s.MustGetHostZone(HostChainId), "legacy response cannot update accounting")
	s.Require().Equal([]icqtypes.Query{freshQuery}, s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "fresh query remains pending")
}

func (s *KeeperTestSuite) TestSubmitDelegationICQ_MissingDelegationAddress() {
	hostZone, _ := s.SetupSubmitDelegationICQ()

	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, ValAddress)
	s.Require().ErrorContains(err, "no delegation address found for")
}

func (s *KeeperTestSuite) TestSubmitDelegationICQ_MissingConnectionId() {
	hostZone, _ := s.SetupSubmitDelegationICQ()

	hostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := s.App.StakeibcKeeper.SubmitDelegationICQ(s.Ctx, hostZone, ValAddress)
	s.Require().ErrorContains(err, "connection-id cannot be empty")
}
