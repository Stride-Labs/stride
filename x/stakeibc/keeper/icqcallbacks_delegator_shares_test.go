package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type DelegatorSharesICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type DelegatorSharesICQCallbackTestCase struct {
	valIndexQueried          int
	hostZone                 types.HostZone
	validArgs                DelegatorSharesICQCallbackArgs
	numShares                sdk.Dec
	slashPercentage          sdk.Dec
	expectedDelegationAmount sdkmath.Int
	expectedSlashAmount      sdkmath.Int
	expectedWeight           uint64
	sharesToTokensRate       sdk.Dec
	retryTimeoutDuration     time.Duration
}

// Mocks the query response that's returned from an ICQ for the number of shares for a given validator/delegator pair
func (s *KeeperTestSuite) CreateDelegatorSharesQueryResponse(valAddress string, shares sdk.Dec) []byte {
	delegation := stakingtypes.Delegation{
		ValidatorAddress: valAddress,
		DelegatorAddress: "cosmos_DELEGATION",
		Shares:           shares,
	}
	delegationBz := s.App.RecordsKeeper.Cdc.MustMarshal(&delegation)
	return delegationBz
}

func (s *KeeperTestSuite) SetupDelegatorSharesICQCallback() DelegatorSharesICQCallbackTestCase {
	// Setting this up to initialize the coordinator for the block time
	s.CreateTransferChannel(HostChainId)

	valIndexQueried := 1
	tokensBeforeSlash := sdkmath.NewInt(1000)
	sharesToTokensRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := sdk.NewDec(1900)

	// 1900 shares * 0.5 sharesToTokens rate = 950 tokens
	// 1000 tokens - 950 token = 50 tokens slashed
	// 50 slash tokens / 1000 initial tokens = 5% slash
	expectedTokensAfterSlash := sdkmath.NewInt(950)
	expectedSlashAmount := tokensBeforeSlash.Sub(expectedTokensAfterSlash)
	slashPercentage := sdk.MustNewDecFromStr("0.05")
	weightBeforeSlash := uint64(20)
	expectedWeightAfterSlash := uint64(19)
	totalDelegation := sdkmath.NewInt(10_000)

	s.Require().Equal(numShares, sdk.NewDecFromInt(expectedTokensAfterSlash.Mul(sdkmath.NewInt(2))), "tokens, shares, and sharesToTokens rate aligned")
	s.Require().Equal(slashPercentage, sdk.NewDecFromInt(expectedSlashAmount).Quo(sdk.NewDecFromInt(tokensBeforeSlash)), "expected slash percentage")
	s.Require().Equal(slashPercentage, sdk.NewDec(int64(weightBeforeSlash-expectedWeightAfterSlash)).Quo(sdk.NewDec(int64(weightBeforeSlash))), "weight reduction")

	hostZone := types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: totalDelegation,
		Validators: []*types.Validator{
			// This validator isn't being queried
			{
				Name:       "val1",
				Address:    "valoper1",
				Weight:     1,
				Delegation: sdkmath.ZeroInt(),
			},
			// This is the validator in question
			{
				Name:                        "val2",
				Address:                     ValAddress,
				SharesToTokensRate:          sharesToTokensRate,
				Delegation:                  tokensBeforeSlash,
				Weight:                      weightBeforeSlash,
				SlashQueryInProgress:        true,
				DelegationChangesInProgress: 0,
			},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, numShares)

	// Create callback data
	callbackDataBz, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: tokensBeforeSlash,
	})
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Set the timeout timestamp to be 1 minute after the block time, and
	// the timeout duration to be 5 minutes
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Ctx = s.Ctx.WithBlockTime(blockTime)
	timeoutDuration := time.Minute
	timeoutTimestamp := uint64(blockTime.Add(timeoutDuration).UnixNano())

	// Create the query that represents the ICQ in flight
	query := icqtypes.Query{
		Id:               "query-1",
		ChainId:          HostChainId,
		ConnectionId:     ibctesting.FirstConnectionID,
		QueryType:        icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		CallbackData:     callbackDataBz,
		CallbackId:       keeper.ICQCallbackID_Delegation,
		CallbackModule:   types.ModuleName,
		TimeoutDuration:  timeoutDuration,
		TimeoutTimestamp: timeoutTimestamp,
		RequestSent:      true,
		TimeoutPolicy:    icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)

	return DelegatorSharesICQCallbackTestCase{
		valIndexQueried: valIndexQueried,
		validArgs: DelegatorSharesICQCallbackArgs{
			query:        query,
			callbackArgs: queryResponse,
		},
		hostZone:                 hostZone,
		numShares:                numShares,
		slashPercentage:          slashPercentage,
		expectedDelegationAmount: expectedTokensAfterSlash,
		expectedSlashAmount:      expectedSlashAmount,
		expectedWeight:           expectedWeightAfterSlash,
		sharesToTokensRate:       sharesToTokensRate,
		retryTimeoutDuration:     timeoutDuration,
	}
}

// Helper function to check if the query was resubmitted in the event that it overlapped an ICA
func (s *KeeperTestSuite) CheckQueryWasResubmitted(tc DelegatorSharesICQCallbackTestCase, hostZone types.HostZone) {
	// After removing the original query, there should be only one query left
	s.App.InterchainqueryKeeper.DeleteQuery(s.Ctx, "query-1")
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "one query expected after re-submission")

	actualQuery := queries[0]
	expectedQuery := tc.validArgs.query

	s.Require().Equal(HostChainId, actualQuery.ChainId, "query chain id")
	s.Require().Equal(ibctesting.FirstConnectionID, actualQuery.ConnectionId, "query connection-id")
	s.Require().Equal(icqtypes.STAKING_STORE_QUERY_WITH_PROOF, actualQuery.QueryType, "query type")

	s.Require().Equal(expectedQuery.CallbackModule, actualQuery.CallbackModule, "query callback module")
	s.Require().Equal(expectedQuery.CallbackId, actualQuery.CallbackId, "query callback id")
	s.Require().Equal(expectedQuery.CallbackData, actualQuery.CallbackData, "query callback data")

	expectedTimeout := s.Ctx.BlockTime().UnixNano() + (tc.retryTimeoutDuration.Nanoseconds())
	s.Require().Equal(expectedTimeout, int64(actualQuery.TimeoutTimestamp), "query timeout timestamp")

	// Confirm the validator still has a query flagged as in progress
	validator := hostZone.Validators[tc.valIndexQueried]
	s.Require().True(validator.SlashQueryInProgress, "slash query is progress")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Successful() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Callback
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback error")

	// Confirm the staked balance was decreased on the host
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.expectedSlashAmount.Int64(), tc.hostZone.TotalDelegations.Sub(hostZone.TotalDelegations).Int64(), "staked bal slash")

	// Confirm the validator's weight and delegation amount were reduced
	validator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(tc.expectedWeight, validator.Weight, "validator weight")
	s.Require().Equal(tc.expectedDelegationAmount.Int64(), validator.Delegation.Int64(), "validator delegation amount")

	// Confirm the validator query is no longer in progress
	s.Require().False(validator.SlashQueryInProgress, "slash query in progress")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Retry_DelegationChange() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Change the validator's delegation in the internal record keeping
	// to make it look as if a delegation ICA landed while the query was in flight
	hostZone := tc.hostZone
	initialDelegation := hostZone.Validators[tc.valIndexQueried].Delegation.Add(sdk.NewInt(100))
	hostZone.Validators[tc.valIndexQueried].Delegation = initialDelegation
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Callback
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "no error expected during delegator shares callback")

	// Confirm the validator's delegation was not modified
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(initialDelegation.Int64(), hostZone.Validators[tc.valIndexQueried].Delegation.Int64(), "validator delegation")

	// Confirm the query was resubmitted
	// The new delegation amount should be stored in the callback data
	callbackDataBz, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: initialDelegation,
	})
	s.Require().NoError(err, "no error expected when marshalling callback data")
	tc.validArgs.query.CallbackData = callbackDataBz

	s.CheckQueryWasResubmitted(tc, hostZone)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Retry_DelegationICAInProgress() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the validator's delegation change ICA counter to show a change is in progress
	initialHostZone := tc.hostZone
	initialHostZone.Validators[tc.valIndexQueried].DelegationChangesInProgress = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, initialHostZone)

	// Callback
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "no error expected during delegator shares callback")

	// Confirm the validator's delegation was not modified
	actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	initialDelegation := initialHostZone.Validators[tc.valIndexQueried].Delegation
	s.Require().Equal(initialDelegation.Int64(), actualHostZone.Validators[tc.valIndexQueried].Delegation.Int64(), "validator delegation")

	// Confirm the query was resubmitted
	s.CheckQueryWasResubmitted(tc, actualHostZone)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_RetryFailure() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Change the validator's delegation in the internal record keeping
	// to make it look as if a delegation ICA landed while the query was in flight
	hostZone := tc.hostZone
	initialDelegation := hostZone.Validators[tc.valIndexQueried].Delegation.Add(sdk.NewInt(100))
	hostZone.Validators[tc.valIndexQueried].Delegation = initialDelegation
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Remove the query connection ID so the retry attempt fails
	invalidQuery := tc.validArgs.query
	invalidQuery.ConnectionId = ""

	// Trigger the callback - this should attempt to retry the query
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to resubmit delegator shares query: failed to retry query")
}

func (s *KeeperTestSuite) checkStateIfValidatorNotSlashed(tc DelegatorSharesICQCallbackTestCase) {
	// Confirm validator on host zone did not update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	initialValidator := tc.hostZone.Validators[tc.valIndexQueried]
	finalValidator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(initialValidator.Weight, finalValidator.Weight, "validator weight should not have updated")
	s.Require().Equal(initialValidator.Delegation, finalValidator.Delegation, "validator delegation amount should not have updated")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_HostZoneNotFound() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Set an incorrect host zone in the query
	badQuery := tc.validArgs.query
	badQuery.ChainId = "fake_host_zone"

	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_InvalidCallbackArgs() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Submit callback with invalid callback args (so that it can't unmarshal into a validator)
	invalidArgs := []byte("random bytes")
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "unable to unmarshal delegator shares query response into Delegation type")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_ValidatorNotFound() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the callback args to contain a validator address that doesn't exist
	badCallbackArgs := s.CreateDelegatorSharesQueryResponse("fake_val", sdk.NewDec(1000)) // 1000 is aribtrary
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no registered validator for address (fake_val): validator not found")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_NoSlashOccurred() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the delegator shares query response so that it shows that there was no slash
	// shares_after_slash = (100% - slash_percentage) * share_if_not_slashed
	//    => share_if_not_slashed = shares_after_slash / (100% - slash_percentage)
	validatorSharesIfNotSlashed := tc.numShares.Quo(sdk.OneDec().Sub(tc.slashPercentage))
	valAddress := tc.hostZone.Validators[tc.valIndexQueried].Address
	queryResponse := s.CreateDelegatorSharesQueryResponse(valAddress, validatorSharesIfNotSlashed)

	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback callback error")

	s.checkStateIfValidatorNotSlashed(tc)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_InvalidNumTokens() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the delegator shares query response so that it shows that there are more tokens delegated
	// than were tracked in state (which shouldn't be possible)
	// Any large number of shares will work here so we'll use 10_000
	valAddress := tc.hostZone.Validators[tc.valIndexQueried].Address
	numShares := sdk.NewDec(10_000)

	badCallbackArgs := s.CreateDelegatorSharesQueryResponse(valAddress, numShares)
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)

	expectedErrMsg := "tokens returned from query is greater than the Delegation: invalid request"
	s.Require().ErrorContains(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_WeightOverfow() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the validator weight to max int so it overflows when casted
	hostZone := tc.hostZone
	validator := hostZone.Validators[tc.valIndexQueried]
	validator.Weight = math.MaxUint64
	hostZone.Validators[tc.valIndexQueried] = validator
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	expectedErrMsg := `unable to convert validator weight to int64, err: overflow: `
	expectedErrMsg += `unable to cast \d+ of type uint64 to int64: unable to cast to safe cast int`
	s.Require().Regexp(expectedErrMsg, err.Error())
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_PrecisionError() {
	tc := s.SetupDelegatorSharesICQCallback()
	initialValidator := tc.hostZone.Validators[tc.valIndexQueried]

	// Update the delegator shares query response so that it shows that there are 5 more tokens delegated
	// than were tracked in state
	// This should be interpretted as a precision error and our record keeping should be adjusted
	precisionErrorTokens := sdk.NewInt(5)
	precisionErrorShares := sdk.NewDecFromInt(precisionErrorTokens).Quo(tc.sharesToTokensRate)
	sharesBeforeSlash := sdk.NewDecFromInt(initialValidator.Delegation).Quo(tc.sharesToTokensRate)

	queryShares := sharesBeforeSlash.Add(precisionErrorShares)
	callbackArgs := s.CreateDelegatorSharesQueryResponse(initialValidator.Address, queryShares)
	err := keeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm host zone and validator were updated
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	expectedTotalDelegation := tc.hostZone.TotalDelegations.Add(precisionErrorTokens)
	s.Require().Equal(expectedTotalDelegation.Int64(), hostZone.TotalDelegations.Int64(), "host zone staked balance")

	validator := hostZone.Validators[tc.valIndexQueried]
	expectedValDelegation := tc.hostZone.Validators[tc.valIndexQueried].Delegation.Add(precisionErrorTokens)
	s.Require().Equal(expectedValDelegation.Int64(), validator.Delegation.Int64(), "validator delegation amount")
}
