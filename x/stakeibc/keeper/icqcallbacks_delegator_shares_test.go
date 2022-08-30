package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
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
	valIndex                 int
	initialState             DelegatorSharesICQCallbackState
	validArgs                DelegatorSharesICQCallbackArgs
	expectedDelegationAmount uint64
	expectedSlashAmount      uint64
	expectedWeight           uint64
}

// Mocks the query response that's returned from an ICQ for the number of shares for a given validator/delegator pair
func (s *KeeperTestSuite) CreateDelegatorSharesQueryResponse(valAddress string, shares uint64) []byte {
	delegation := stakingtypes.Delegation{
		ValidatorAddress: valAddress,
		DelegatorAddress: "cosmos_DELEGATION",
		Shares:           sdk.NewDec(int64(shares)),
	}
	delegationBz := s.App.RecordsKeeper.Cdc.MustMarshal(&delegation)
	return delegationBz
}

func (s *KeeperTestSuite) SetupDelegatorSharesICQCallback() DelegatorSharesICQCallbackTestCase {
	// Setting this up to initialize the coordinator for the block time
	s.CreateTransferChannel(HostChainId)

	valAddress := "valoper2"
	valIndex := 1
	tokensBeforeSlash := uint64(1000)
	internalExchangeRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := uint64(1900)

	// 1900 shares * 0.5 exchange rate = 950 tokens
	// 1000 tokens - 950 token = 50 tokens slashed
	// 50 slash tokens / 1000 initial tokens = 5% slash
	expectedTokensAfterSlash := uint64(950)
	expectedSlashAmount := tokensBeforeSlash - expectedTokensAfterSlash
	weightBeforeSlash := uint64(20)
	expectedWeightAfterSlash := uint64(19)
	stakedBal := uint64(10_000)

	s.Require().Equal(numShares, expectedTokensAfterSlash*2, "tokens, shares, and exchange rate aligned")
	s.Require().Equal(0.05, float64(expectedSlashAmount)/float64(tokensBeforeSlash), "expected slash percentage")
	s.Require().Equal(0.05, float64(weightBeforeSlash-expectedWeightAfterSlash)/float64(weightBeforeSlash), "weight reduction")

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

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)

	queryResponse := s.CreateDelegatorSharesQueryResponse(valAddress, numShares)

	return DelegatorSharesICQCallbackTestCase{
		valIndex: valIndex,
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
		expectedDelegationAmount: expectedTokensAfterSlash,
		expectedSlashAmount:      expectedSlashAmount,
		expectedWeight:           expectedWeightAfterSlash,
	}
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Successful() {
	tc := s.SetupDelegatorSharesICQCallback()

	// Callback
	err := stakeibckeeper.DelegatorSharesCallback(s.App.StakeibcKeeper, s.Ctx(), tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err, "delegator shares callback error")

	// Confirm the staked balance was decreased on the host
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), tc.initialState.hostZone.ChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.expectedSlashAmount, tc.initialState.hostZone.StakedBal-hostZone.StakedBal)

	// Confirm the validator's weight and delegation amount were reduced
	validator := hostZone.Validators[tc.valIndex]
	s.Require().Equal(tc.expectedWeight, validator.Weight, "validator weight")
	s.Require().Equal(tc.expectedDelegationAmount, validator.DelegationAmt, "validator delegation amount")
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_InvalidCallbackArgs() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_BufferWindowError() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_OutsideBufferWindow() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_ValidatorNotFound() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_EpochNotFound() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_ExchangeRateNotFound() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_NoSlashOccurred() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_SlashOverfow() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_WeightOverfow() {

}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_SlashGtTenPercent() {

}
