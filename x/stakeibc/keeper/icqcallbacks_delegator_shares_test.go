package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
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
	initialState      DelegatorSharesICQCallbackState
	validArgs         DelegatorSharesICQCallbackArgs
	expectedNumTokens int64
}

// Mocks the query response that's returned from an ICQ for the number of shares for a given validator/delegator pair
func (s *KeeperTestSuite) CreateDelegatorSharesQueryResponse(valAddress string, shares int64) []byte {
	delegation := stakingtypes.Delegation{
		ValidatorAddress: valAddress,
		DelegatorAddress: "cosmos_DELEGATION",
		Shares:           sdk.NewDec(shares),
	}
	delegationBz := s.App.RecordsKeeper.Cdc.MustMarshal(&delegation)
	return delegationBz
}

func (s *KeeperTestSuite) SetupDelegatorSharesICQCallback() DelegatorSharesICQCallbackTestCase {
	currentEpoch := uint64(1)

	// With an internal validator exchange rate of 0.5 and 2000 shares for this delegator
	// We'd expect that to translate to 1000 tokens for the delegator/validator pair
	valAddress := "valoper2"
	internalExchangeRate := sdk.NewDec(1).Quo(sdk.NewDec(2)) // 0.5
	numShares := int64(2000)
	expectedNumTokens := int64(1000)
	delegationAmount := uint64(950) // this means the validator was slashed 5%
	weight := uint64(10)

	hostZone := stakeibctypes.HostZone{
		ChainId: HostChainId,
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
				DelegationAmt: delegationAmount,
				Weight:        weight,
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
		expectedNumTokens: expectedNumTokens,
	}
}

func (s *KeeperTestSuite) TestDelegatorSharesCallback_Successful() {

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
