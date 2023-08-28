package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type InstantiateOracleCallbackTestCase struct {
	ContractAddress     string
	ValidCallbackArgs   []byte
	ValidICAMsgResponse [][]byte
	ValidAckResponse    icacallbacktypes.AcknowledgementResponse
}

func (s *KeeperTestSuite) SetupTestInstantiateOracleCallback() InstantiateOracleCallbackTestCase {
	// Store oracle
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, types.Oracle{
		ChainId: HostChainId,
		Active:  false,
	})

	// Confirm it was stored
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be in the store during setup")

	// Build ack response
	contractAddress := "contract_address"
	icaResponse := types.MsgInstantiateContractResponse{
		Address: contractAddress,
	}
	icaResponseBz, err := proto.Marshal(&icaResponse)
	s.Require().NoError(err, "no error expected when marshalling contract response")
	icaMsgResponse := [][]byte{icaResponseBz}

	// Build callback data
	callbackArgs := types.InstantiateOracleCallback{
		OracleChainId: HostChainId,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err, "no error expected when marshalling callback args")

	return InstantiateOracleCallbackTestCase{
		ContractAddress:     contractAddress,
		ValidCallbackArgs:   callbackArgsBz,
		ValidICAMsgResponse: icaMsgResponse,
		ValidAckResponse: icacallbacktypes.AcknowledgementResponse{
			Status:       icacallbacktypes.AckResponseStatus_SUCCESS,
			MsgResponses: icaMsgResponse,
		},
	}
}

// Helper function to check the state after the callback
func (s *KeeperTestSuite) checkStateAfterInstantiateCallback(success bool, expectedContractAddress string) {
	oracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should have been found after callback")

	expectedActive := success // if successful, it should be active

	s.Require().Equal(expectedActive, oracle.Active, "oracle active field")
	s.Require().Equal(oracle.ContractAddress, expectedContractAddress, "oracle contract address")
}

// Function to call the callback, and check the state depending on whether it was successful
func (s *KeeperTestSuite) executeCallbackAndCheckState(
	tc InstantiateOracleCallbackTestCase,
	ackStatus icacallbacktypes.AckResponseStatus,
	callbackArgs []byte,
) {
	// Build the ack response object with the provided status
	ackResponse := icacallbacktypes.AcknowledgementResponse{
		Status:       ackStatus,
		MsgResponses: tc.ValidICAMsgResponse,
	}

	// The callback should not throw an error in these cases (even in the event of a timeout/ack failure)
	err := s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &ackResponse, callbackArgs)
	s.Require().Nil(err, "no error expected during callback")

	// If the ack was a success, check that the contract address was updated
	if ackStatus == icacallbacktypes.AckResponseStatus_SUCCESS {
		success := true
		expectedContractAddress := tc.ContractAddress
		s.checkStateAfterInstantiateCallback(success, expectedContractAddress)
	} else {
		// Otherwise, during timeout / ack failure, check that no state was changed
		success := false
		expectedContractAddress := ""
		s.checkStateAfterInstantiateCallback(success, expectedContractAddress)
	}
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_AckSuccess() {
	tc := s.SetupTestInstantiateOracleCallback()
	s.executeCallbackAndCheckState(tc, icacallbacktypes.AckResponseStatus_SUCCESS, tc.ValidCallbackArgs)
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_AckTimeout() {
	tc := s.SetupTestInstantiateOracleCallback()
	s.executeCallbackAndCheckState(tc, icacallbacktypes.AckResponseStatus_TIMEOUT, tc.ValidCallbackArgs)
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_AckFailure() {
	tc := s.SetupTestInstantiateOracleCallback()
	s.executeCallbackAndCheckState(tc, icacallbacktypes.AckResponseStatus_FAILURE, tc.ValidCallbackArgs)
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_UnmarshalCallbackFailure() {
	tc := s.SetupTestInstantiateOracleCallback()

	// Calling the callback with invalid callback args should fail
	invalidArgs := []byte{1, 2, 3}
	err := s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &tc.ValidAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "unable to unmarshal instantiate oracle callback")
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_OracleNotFound() {
	tc := s.SetupTestInstantiateOracleCallback()

	// Remove the oracle
	s.App.ICAOracleKeeper.RemoveOracle(s.Ctx, HostChainId)

	// Call the callback - should fail
	err := s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &tc.ValidAckResponse, tc.ValidCallbackArgs)
	s.Require().ErrorContains(err, "oracle not found")
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_NoMessagesInICAResponse() {
	tc := s.SetupTestInstantiateOracleCallback()

	// Calling the callback with no messages in the ICA response should fail
	invalidAckResponse := tc.ValidAckResponse
	invalidAckResponse.MsgResponses = [][]byte{}

	err := s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &invalidAckResponse, tc.ValidCallbackArgs)
	s.Require().ErrorContains(err, "tx response from CW contract instantiation should have 1 message (0 found)")
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_UnmarshalICAResponseFailure() {
	tc := s.SetupTestInstantiateOracleCallback()

	// Calling the callback with an invalid ack response should fail
	invalidAckResponse := tc.ValidAckResponse
	invalidAckResponse.MsgResponses = [][]byte{{1, 2, 3}}

	err := s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &invalidAckResponse, tc.ValidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal instantiate contract response")
}

func (s *KeeperTestSuite) TestInstantiateOracleCallback_NoContractAddressInICAResponse() {
	tc := s.SetupTestInstantiateOracleCallback()

	// Create an ack response that does not contain a contract address
	responseWithNoContract := types.MsgInstantiateContractResponse{}
	responseWithNoContractBz, err := proto.Marshal(&responseWithNoContract)
	s.Require().NoError(err, "no error expected when marshalling contract response")

	invalidAckResponse := tc.ValidAckResponse
	invalidAckResponse.MsgResponses = [][]byte{responseWithNoContractBz}

	err = s.App.ICAOracleKeeper.InstantiateOracleCallback(s.Ctx, channeltypes.Packet{}, &invalidAckResponse, tc.ValidCallbackArgs)
	s.Require().ErrorContains(err, "response from CW contract instantiation ICA does not contain a contract address")
}
