package keeper_test

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
	expectedDelegationAmount sdk.Int
	expectedSlashAmount      sdk.Int
	expectedWeight           uint64
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
	tokensBeforeSlash := sdk.NewInt(1000)
	internalExchangeRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := sdk.NewDec(1900)

	// 1900 shares * 0.5 exchange rate = 950 tokens
	// 1000 tokens - 950 token = 50 tokens slashed
	// 50 slash tokens / 1000 initial tokens = 5% slash
	expectedTokensAfterSlash := sdk.NewInt(950)
	expectedSlashAmount := tokensBeforeSlash.Sub(expectedTokensAfterSlash) 
	slashPercentage := sdk.MustNewDecFromStr("0.05")
	weightBeforeSlash := uint64(20)
	expectedWeightAfterSlash := uint64(19)
	stakedBal := sdk.NewInt(10_000)

	s.Require().Equal(numShares, sdk.NewDecFromInt(expectedTokensAfterSlash.Mul(sdk.NewInt(2))), "tokens, shares, and exchange rate aligned")
	s.Require().Equal(slashPercentage, sdk.NewDecFromInt(expectedSlashAmount).Quo(sdk.NewDecFromInt(tokensBeforeSlash)), "expected slash percentage")
	s.Require().Equal(slashPercentage, sdk.NewDec(int64(weightBeforeSlash-expectedWeightAfterSlash)).Quo(sdk.NewDec(int64(weightBeforeSlash))), "weight reduction")

	currentEpoch := uint64(1)
	hostZone := stakeibctypes.HostZone{
		ChainId:   HostChainId,
		StakedBal: stakedBal,
		Validators: []*stakeibctypes.Validator{
			// This validator isn't being queried
			{
				Name:    "val1",
				Address: "valoper1",
				Weight:  1,
				DelegationAmt: sdk.ZeroInt(),
			},
			// This is the validator in question
			{
				Name:    "val2",
				Address: valAddress,
				InternalExchangeRate: &stakeibctypes.ValidatorExchangeRate{
					InternalTokensToSharesRate: internalExchangeRate,
					EpochNumber:                currentEpoch,
				},
				DelegationAmt: tokensBeforeSlash,
				Weight:        weightBeforeSlash,
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

	return DelegatorSharesICQCallbackTestCase{
		valIndexQueried: valIndexQueried,
		initialState: DelegatorSharesICQCallbackState{
			hostZone:           hostZone,
			strideEpochTracker: strideEpochTracker,
		},
		validArgs: DelegatorSharesICQCallbackArgs{
			query: icqtypes.Query{
				ChainId: HostChainId,
			},
			callbackArgs: queryResponse,
		},
		numShares:                numShares,
		slashPercentage:          slashPercentage,
		expectedDelegationAmount: expectedTokensAfterSlash,
		expectedSlashAmount:      expectedSlashAmount,
		expectedWeight:           expectedWeightAfterSlash,
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
	s.Require().Equal(tc.expectedSlashAmount, tc.initialState.hostZone.StakedBal.Sub(hostZone.StakedBal))

	// Confirm the validator's weight and delegation amount were reduced
	validator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(tc.expectedWeight, validator.Weight, "validator weight")
	s.Require().Equal(tc.expectedDelegationAmount, validator.DelegationAmt, "validator delegation amount")
}

func (s *KeeperTestSuite) checkStateIfValidatorNotSlashed(tc DelegatorSharesICQCallbackTestCase) {
	// Confirm validator on host zone did not update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")

	initialValidator := tc.initialState.hostZone.Validators[tc.valIndexQueried]
	finalValidator := hostZone.Validators[tc.valIndexQueried]
	s.Require().Equal(initialValidator.Weight, finalValidator.Weight, "validator weight should not have updated")
	s.Require().Equal(initialValidator.DelegationAmt, finalValidator.DelegationAmt, "validator delegation amount should not have updated")
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

	expectedErrMsg := "unable to unmarshal queried delegation info for zone GAIA, "
	expectedErrMsg += "err: unexpected EOF: unable to marshal data structure"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_BufferWindowError() {
	tc := s.SetupDelegatorSharesICQCallback()

	// update epoch tracker so that we're in the middle of an epoch
	epochTracker := tc.initialState.strideEpochTracker
	epochTracker.Duration = 0 // duration of 0 will make the epoch start time equal to the epoch end time

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)

	s.Require().ErrorContains(err, "unable to determine if ICQ callback is inside buffer window")
	s.Require().ErrorContains(err, "current block time")
	s.Require().ErrorContains(err, "not within current epoch")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_OutsideBufferWindow() {
	tc := s.SetupDelegatorSharesICQCallback()

	// update epoch tracker so that we're in the middle of an epoch
	epochTracker := tc.initialState.strideEpochTracker
	epochTracker.Duration = 10_000_000_000                                                         // 10 second epochs
	epochTracker.NextEpochStartTime = uint64(s.Coordinator.CurrentTime.UnixNano() + 5_000_000_000) // epoch ends in 5 second

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	// In this case, we should return success instead of error, but we should exit early before updating the validator's state
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback callback error")

	s.checkStateIfValidatorNotSlashed(tc)
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
	s.Require().EqualError(err, "DelegationCallback: validator (valoper2) internal exchange rate has not been updated this epoch (epoch #2): invalid request")
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
	// that were tracked in state (which shouldn't be possible)
	// Any large number of shares will work here so we'll use 10_000
	valAddress := tc.initialState.hostZone.Validators[tc.valIndexQueried].Address
	numShares := sdk.NewDec(10_000)

	badCallbackArgs := s.CreateDelegatorSharesQueryResponse(valAddress, numShares)
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)

	expectedErrMsg := "DelegationCallback: Validator (valoper2) tokens returned from query is greater than the DelegationAmt: invalid request"
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
	expectedErrMsg := "DelegationCallback: Validator (valoper2) slashed but ABORTING update, "
	expectedErrMsg += "slash is greater than 0.10 (0.200000000000000000): slash is greater than 10 percent"
	s.Require().EqualError(err, expectedErrMsg)
}
