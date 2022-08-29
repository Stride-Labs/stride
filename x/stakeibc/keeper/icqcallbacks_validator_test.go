package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

var HostChainId = "GAIA"

type ValidatorICQCallbackState struct {
	hostZone stakeibctypes.HostZone
}

type ValidatorICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type ValidatorICQCallbackTestCase struct {
	initialState         ValidatorICQCallbackState
	validArgs            ValidatorICQCallbackArgs
	valAddress           string
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

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		Duration:           10_000_000_000,                                               // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 1_000_000_000), // epoch ends in 1 second
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)

	queryResponse := s.CreateValidatorQueryResponse(valAddress, numTokens, numShares)

	return ValidatorICQCallbackTestCase{
		initialState: ValidatorICQCallbackState{
			hostZone: hostZone,
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

	err := stakeibckeeper.ValidatorExchangeRateCallback(s.App.StakeibcKeeper, s.Ctx(), tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "called valdiator exchange rate callback")

	// Confirm validator's exchange rate was update
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.expectedExchangeRate, hostZone.Validators[tc.valIndexQueried].InternalExchangeRate.InternalTokensToSharesRate,
		"validator exchange rate updated")
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_InvalidCallbackArgs() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_BufferWindowError() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_OutsideBufferWindow() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_ValidatorNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_EpochNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_DelegationQueryFailed() {

}
