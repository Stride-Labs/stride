package keeper_test

import (
	"math"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck

	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type DelegatorSharesICQCallbackState struct {
	hostZone           stakeibctypes.HostZone
	strideEpochTracker stakeibctypes.EpochTracker
}

type DelegatorSharesICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type DelegatorSharesICQCallbackTestCase struct {
	valIndexQueried          int
	initialState             DelegatorSharesICQCallbackState
	validArgs                DelegatorSharesICQCallbackArgs
	numShares                sdk.Dec
	slashPercentage          sdk.Dec
	expectedDelegationAmount sdkmath.Int
	expectedSlashAmount      sdkmath.Int
	expectedWeight           uint64
	exchangeRate             sdk.Dec
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

	valAddress := "valoper2"
	valIndexQueried := 1
	tokensBeforeSlash := sdkmath.NewInt(1000)
	internalExchangeRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := sdk.NewDec(1900)

	// 1900 shares * 0.5 exchange rate = 950 tokens
	// 1000 tokens - 950 token = 50 tokens slashed
	// 50 slash tokens / 1000 initial tokens = 5% slash
	expectedTokensAfterSlash := sdkmath.NewInt(950)
	expectedSlashAmount := tokensBeforeSlash.Sub(expectedTokensAfterSlash)
	slashPercentage := sdk.MustNewDecFromStr("0.05")
	weightBeforeSlash := uint64(20)
	expectedWeightAfterSlash := uint64(19)
	totalDelegation := sdkmath.NewInt(10_000)

	s.Require().Equal(numShares, sdk.NewDecFromInt(expectedTokensAfterSlash.Mul(sdkmath.NewInt(2))), "tokens, shares, and exchange rate aligned")
	s.Require().Equal(slashPercentage, sdk.NewDecFromInt(expectedSlashAmount).Quo(sdk.NewDecFromInt(tokensBeforeSlash)), "expected slash percentage")
	s.Require().Equal(slashPercentage, sdk.NewDec(int64(weightBeforeSlash-expectedWeightAfterSlash)).Quo(sdk.NewDec(int64(weightBeforeSlash))), "weight reduction")

	currentEpoch := uint64(1)
	hostZone := stakeibctypes.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: totalDelegation,
		Validators: []*stakeibctypes.Validator{
			// This validator isn't being queried
			{
				Name:       "val1",
				Address:    "valoper1",
				Weight:     1,
				Delegation: sdkmath.ZeroInt(),
			},
			// This is the validator in question
			{
				Name:    "val2",
				Address: valAddress,
				InternalExchangeRate: &stakeibctypes.ValidatorExchangeRate{
					InternalTokensToSharesRate: internalExchangeRate,
					EpochNumber:                currentEpoch,
				},
				Delegation: tokensBeforeSlash,
				Weight:     weightBeforeSlash,
			},
		},
	}

	// This will make the current time 90% through the epoch
	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           10_000_000_000,                                               // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 1_000_000_000), // epoch ends in 1 second
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	queryResponse := s.CreateDelegatorSharesQueryResponse(valAddress, numShares)

	// Create callback data
	timeoutDuration := time.Hour
	callbackDataBz, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: tokensBeforeSlash,
		TimeoutDuration:            timeoutDuration,
	})
	s.Require().NoError(err, "no error expected when marshalling callback data")

	return DelegatorSharesICQCallbackTestCase{
		valIndexQueried: valIndexQueried,
		initialState: DelegatorSharesICQCallbackState{
			hostZone:           hostZone,
			strideEpochTracker: strideEpochTracker,
		},
		validArgs: DelegatorSharesICQCallbackArgs{
			query: icqtypes.Query{
				ChainId:        HostChainId,
				ConnectionId:   ibctesting.FirstConnectionID,
				QueryType:      icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
				CallbackData:   callbackDataBz,
				CallbackId:     keeper.ICQCallbackID_Delegation,
				CallbackModule: types.ModuleName,
			},
			callbackArgs: queryResponse,
		},
		numShares:                numShares,
		slashPercentage:          slashPercentage,
		expectedDelegationAmount: expectedTokensAfterSlash,
		expectedSlashAmount:      expectedSlashAmount,
		expectedWeight:           expectedWeightAfterSlash,
		exchangeRate:             internalExchangeRate,
		retryTimeoutDuration:     timeoutDuration,
	}
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Successful() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Callback
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback error")

	// Confirm the staked balance was decreased on the host
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.expectedSlashAmount.Int64(), tc.initialState.hostZone.TotalDelegations.Sub(hostZone.TotalDelegations).Int64(), "staked bal slash")

	// Confirm the validator's weight and delegation amount were reduced
	validator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(tc.expectedWeight, validator.Weight, "validator weight")
	s.Require().Equal(tc.expectedDelegationAmount.Int64(), validator.Delegation.Int64(), "validator delegation amount")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Retry() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Change the validator's delegation in the internal record keeping
	// to make it look as if a delegation ICA landed while the query was in flight
	hostZone := tc.initialState.hostZone
	initialDelegation := hostZone.Validators[tc.valIndexQueried].Delegation.Add(sdk.NewInt(100))
	hostZone.Validators[tc.valIndexQueried].Delegation = initialDelegation
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Callback
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "no error expected during delegator shares callback")

	// Confirm the validator's delegation was not modified
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(initialDelegation.Int64(), hostZone.Validators[tc.valIndexQueried].Delegation.Int64(), "validator delegation")

	// Confirm the query was resubmitted
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
	s.Require().Equal(expectedTimeout, int64(actualQuery.Timeout), "query callback data")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_RetryFailure() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Change the validator's delegation in the internal record keeping
	// to make it look as if a delegation ICA landed while the query was in flight
	hostZone := tc.initialState.hostZone
	initialDelegation := hostZone.Validators[tc.valIndexQueried].Delegation.Add(sdk.NewInt(100))
	hostZone.Validators[tc.valIndexQueried].Delegation = initialDelegation
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Remove the query connection ID so the retry attempt fails
	invalidQuery := tc.validArgs.query
	invalidQuery.ConnectionId = ""

	// Trigger the callback - this should attempt to retry the query
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().ErrorContains(err, "unable to resubmit delegator shares query: connection-id cannot be empty")
}

func (s *KeeperTestSuite) checkStateIfValidatorNotSlashed(tc DelegatorSharesICQCallbackTestCase) {
	// Confirm validator on host zone did not update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")

	initialValidator := tc.initialState.hostZone.Validators[tc.valIndexQueried]
	finalValidator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(initialValidator.Weight, finalValidator.Weight, "validator weight should not have updated")
	s.Require().Equal(initialValidator.Delegation, finalValidator.Delegation, "validator delegation amount should not have updated")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_HostZoneNotFound() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Set an incorrect host zone in the query
	badQuery := tc.validArgs.query
	badQuery.ChainId = "fake_host_zone"

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_InvalidCallbackArgs() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Submit callback with invalid callback args (so that it can't unmarshal into a validator)
	invalidArgs := []byte("random bytes")
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "unable to unmarshal delegator shares query response into Delegation type")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_ValidatorNotFound() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the callback args to contain a validator address that doesn't exist
	badCallbackArgs := s.CreateDelegatorSharesQueryResponse("fake_val", sdk.NewDec(1000)) // 1000 is aribtrary
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no registered validator for address (fake_val): validator not found")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_ExchangeRateNotFound() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Increment the epoch number so that we're in an epoch that has not queried the validator's exchange rate
	epochTracker := tc.initialState.strideEpochTracker
	epochTracker.EpochNumber += 1
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "validator (valoper2) internal exchange rate has not been updated this epoch (epoch #2): invalid request")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_NoSlashOccurred() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the delegator shares query response so that it shows that there was no slash
	// shares_after_slash = (100% - slash_percentage) * share_if_not_slashed
	//    => share_if_not_slashed = shares_after_slash / (100% - slash_percentage)
	validatorSharesIfNotSlashed := tc.numShares.Quo(sdk.OneDec().Sub(tc.slashPercentage))
	valAddress := tc.initialState.hostZone.Validators[tc.valIndexQueried].Address
	queryResponse := s.CreateDelegatorSharesQueryResponse(valAddress, validatorSharesIfNotSlashed)

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback callback error")

	s.checkStateIfValidatorNotSlashed(tc)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_InvalidNumTokens() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the delegator shares query response so that it shows that there are more tokens delegated
	// than were tracked in state (which shouldn't be possible)
	// Any large number of shares will work here so we'll use 10_000
	valAddress := tc.initialState.hostZone.Validators[tc.valIndexQueried].Address
	numShares := sdk.NewDec(10_000)

	badCallbackArgs := s.CreateDelegatorSharesQueryResponse(valAddress, numShares)
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)

	expectedErrMsg := "Validator (valoper2) tokens returned from query is greater than the Delegation: invalid request"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_WeightOverfow() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the validator weight to max int so it overflows when casted
	hostZone := tc.initialState.hostZone
	validator := hostZone.Validators[tc.valIndexQueried]
	validator.Weight = math.MaxUint64
	hostZone.Validators[tc.valIndexQueried] = validator
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	expectedErrMsg := `unable to convert validator weight to int64, err: overflow: `
	expectedErrMsg += `unable to cast \d+ of type uint64 to int64: unable to cast to safe cast int`
	s.Require().Regexp(expectedErrMsg, err.Error())
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_SlashGtTenPercent() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Update the callback args to contain a number of shares that would imply a slash greater than 10%
	valAddress := tc.initialState.hostZone.Validators[tc.valIndexQueried].Address
	badCallbackArgs := s.CreateDelegatorSharesQueryResponse(valAddress, sdk.NewDec(1600))

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)
	expectedErrMsg := "Validator slashed but ABORTING update, slash (0.200000000000000000) is greater than safety threshold (0.100000000000000000): "
	expectedErrMsg += "slash is greater than safety threshold"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_PrecisionError() {
	tc := s.SetupDelegatorSharesICQCallback()
	initialValidator := tc.initialState.hostZone.Validators[tc.valIndexQueried]

	// Update the delegator shares query response so that it shows that there are 5 more tokens delegated
	// than were tracked in state
	// This should be interpretted as a precision error and our record keeping should be adjusted
	precisionErrorTokens := sdk.NewInt(5)
	precisionErrorShares := sdk.NewDecFromInt(precisionErrorTokens).Quo(tc.exchangeRate)
	sharesBeforeSlash := sdk.NewDecFromInt(initialValidator.Delegation).Quo(tc.exchangeRate)

	queryShares := sharesBeforeSlash.Add(precisionErrorShares)
	callbackArgs := s.CreateDelegatorSharesQueryResponse(initialValidator.Address, queryShares)
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm host zone and validator were updated
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")

	expectedTotalDelegation := tc.initialState.hostZone.TotalDelegations.Add(precisionErrorTokens)
	s.Require().Equal(expectedTotalDelegation.Int64(), hostZone.TotalDelegations.Int64(), "host zone staked balance")

	validator := hostZone.Validators[tc.valIndexQueried]
	expectedValDelegation := tc.initialState.hostZone.Validators[tc.valIndexQueried].Delegation.Add(precisionErrorTokens)
	s.Require().Equal(expectedValDelegation.Int64(), validator.Delegation.Int64(), "validator delegation amount")
}
