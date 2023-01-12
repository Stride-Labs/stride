package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type ValidatorICQCallbackState struct {
	hostZone           stakeibctypes.HostZone
	strideEpochTracker stakeibctypes.EpochTracker
}

type ValidatorICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type ValidatorICQCallbackTestCase struct {
	initialState         ValidatorICQCallbackState
	validArgs            ValidatorICQCallbackArgs
	valIndexQueried      int
	expectedExchangeRate sdk.Dec
}

func (s *KeeperTestSuite) CreateValidatorQueryResponse(address string, tokens int64, shares int64) []byte {
	validator := stakingtypes.Validator{
		OperatorAddress: address,
		Tokens:          sdk.NewInt(tokens),
		DelegatorShares: sdk.NewDec(shares),
	}
	validatorBz := s.App.RecordsKeeper.Cdc.MustMarshal(&validator)
	return validatorBz
}

func (s *KeeperTestSuite) SetupValidatorICQCallback() ValidatorICQCallbackTestCase {
	// We don't actually need a transfer channel for this test, but we do need to have IBC support for timeouts
	s.CreateTransferChannel(HostChainId)

	valAddress := "valoper1"
	valIndexQueried := 0 // index in the validators array
	// In this example, the validator has 2000 shares, originally had 2000 tokens,
	// and now has 1000 tokens (after being slashed)
	initialExchangeRate := sdk.NewDec(1)
	numShares := int64(2000)
	numTokens := int64(1000)
	expectedExchangeRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5

	currentEpoch := uint64(2)
	hostZone := stakeibctypes.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		DelegationAccount: &stakeibctypes.ICAAccount{
			Address: "cosmos_DELEGATION",
			Target:  stakeibctypes.ICAAccountType_DELEGATION,
		},
		Validators: []*stakeibctypes.Validator{
			{
				Name:    "val1",
				Address: valAddress,
				InternalExchangeRate: &stakeibctypes.ValidatorExchangeRate{
					InternalTokensToSharesRate: initialExchangeRate,
					EpochNumber:                currentEpoch,
				},
			},
			// This validator isn't being queried
			{
				Name:    "val2",
				Address: "valoper2",
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

	queryResponse := s.CreateValidatorQueryResponse(valAddress, numTokens, numShares)

	return ValidatorICQCallbackTestCase{
		initialState: ValidatorICQCallbackState{
			hostZone:           hostZone,
			strideEpochTracker: strideEpochTracker,
		},
		validArgs: ValidatorICQCallbackArgs{
			query: icqtypes.Query{
				ChainId: HostChainId,
			},
			callbackArgs: queryResponse,
		},
		valIndexQueried:      valIndexQueried,
		expectedExchangeRate: expectedExchangeRate,
	}
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_Successful() {
	tc := s.SetupValidatorICQCallback()

	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator exchange rate callback error")

	// Confirm validator's exchange rate was update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.expectedExchangeRate, hostZone.Validators[tc.valIndexQueried].InternalExchangeRate.InternalTokensToSharesRate,
		"validator exchange rate updated")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_HostZoneNotFound() {
	tc := s.SetupValidatorICQCallback()

	// Set an incorrect host zone in the query
	badQuery := tc.validArgs.query
	badQuery.ChainId = "fake_host_zone"

	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, badQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_InvalidCallbackArgs() {
	tc := s.SetupValidatorICQCallback()

	// Submit callback with invalid callback args (so that it can't unmarshal into a validator)
	invalidArgs := []byte("random bytes")
	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)

	expectedErrMsg := "unable to unmarshal queriedValidator info for zone GAIA, "
	expectedErrMsg += "err: unexpected EOF: unable to marshal data structure"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_BufferWindowError() {
	tc := s.SetupValidatorICQCallback()

	// update epoch tracker so that we're in the middle of an epoch
	epochTracker := tc.initialState.strideEpochTracker
	epochTracker.Duration = 0 // duration of 0 will make the epoch start time equal to the epoch end time

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "unable to determine if ICQ callback is inside buffer window")
	s.Require().ErrorContains(err, "current block time")
	s.Require().ErrorContains(err, "not within current epoch")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_OutsideBufferWindow() {
	tc := s.SetupValidatorICQCallback()

	// update epoch tracker so that we're in the middle of an epoch
	epochTracker := tc.initialState.strideEpochTracker
	epochTracker.Duration = 10_000_000_000                                                         // 10 second epochs
	epochTracker.NextEpochStartTime = uint64(s.Coordinator.CurrentTime.UnixNano() + 5_000_000_000) // epoch ends in 5 second

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	// In this case, we should return success instead of error, but we should exit early before updating the validator's exchange rate
	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "validator exchange rate callback error")

	// Confirm validator's exchange rate did not update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")

	initialExchangeRate := tc.initialState.hostZone.Validators[tc.valIndexQueried].InternalExchangeRate.InternalTokensToSharesRate
	actualExchangeRate := hostZone.Validators[tc.valIndexQueried].InternalExchangeRate.InternalTokensToSharesRate
	s.Require().Equal(actualExchangeRate, initialExchangeRate, "validator exchange rate should not have updated")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_ValidatorNotFound() {
	tc := s.SetupValidatorICQCallback()

	// Update the callback args to contain a validator address that doesn't exist
	badCallbackArgs := s.CreateValidatorQueryResponse("fake_val", 0, 0)
	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no registered validator for address (fake_val): validator not found")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_DelegatorSharesZero() {
	tc := s.SetupValidatorICQCallback()

	// Set the delegator shares to 0, which cause division by zero in `validator.TokensFromShares`
	valAddress := tc.initialState.hostZone.Validators[tc.valIndexQueried].Address
	badCallbackArgs := s.CreateValidatorQueryResponse(valAddress, 1000, 0) // the 1000 is arbitrary, the zero here is what matters
	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, badCallbackArgs, tc.validArgs.query)

	expectedErrMsg := "can't calculate validator internal exchange rate because delegation amount is 0 (validator: valoper1): division by zero"
	s.Require().EqualError(err, expectedErrMsg)
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_DelegationQueryFailed() {
	tc := s.SetupValidatorICQCallback()

	// Remove host zone delegation address so delegation query fails
	badHostZone := tc.initialState.hostZone
	badHostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)

	expectedErrMsg := "ValidatorCallback: failed to query delegation, zone GAIA, err: Zone GAIA is missing a delegation address!: "
	expectedErrMsg += "ICA acccount not found on host zone: failed to submit ICQ"
	s.Require().EqualError(err, expectedErrMsg)
}
