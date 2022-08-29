package keeper_test

import (
	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type ValidatorICQCallbackState struct {
	hostZone stakeibctypes.HostZone
}

type ValidatorICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type ValidatorICQCallbackTestCase struct {
	initialState ValidatorICQCallbackState
	validArgs    ValidatorICQCallbackArgs
}

func (s *KeeperTestSuite) CreateValidatorQueryRequest() []byte {
	return []byte("")
}

func (s *KeeperTestSuite) CreateValidatorQueryResponse() []byte {
	return []byte("")
}

func (s *KeeperTestSuite) SetupValidatorICQCallback() ValidatorICQCallbackTestCase {
	hostZone := stakeibctypes.HostZone{
		ChainId: HostChainId,
		DelegationAccount: &stakeibctypes.ICAAccount{
			Address: "cosmos_DELEGATION",
			Target:  stakeibctypes.ICAAccountType_DELEGATION,
		},
	}

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)

	queryRequest := s.CreateValidatorQueryRequest()
	queryResponse := s.CreateValidatorQueryResponse()

	return ValidatorICQCallbackTestCase{
		initialState: ValidatorICQCallbackState{
			hostZone: hostZone,
		},
		validArgs: ValidatorICQCallbackArgs{
			query: icqtypes.Query{
				Id:      "0",
				ChainId: HostChainId,
				Request: queryRequest,
			},
			callbackArgs: queryResponse,
		},
	}
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_Successful() {
	// Check exchange rate updated on validator
}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_InvalidCallbackArgs() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_BufferWindowError() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_ValidatorNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_EpochNotFound() {

}

func (s *KeeperTestSuite) TestValidatorExchangeRateCallback_DelegationQueryFailed() {

}
