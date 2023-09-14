package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	"github.com/cosmos/gogoproto/proto"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/records/types"
)

var (
	LSMTokenDenom = "cosmosvaloperxxx/42"
)

func (s *KeeperTestSuite) SetupLSMTransferCallback() []byte {
	// we need a valid ibc denom here or the transfer will fail
	prefixedDenom := transfertypes.GetPrefixedDenom(transfertypes.PortID, ibctesting.FirstChannelID, LSMTokenDenom)
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	ibcDenom := denomTrace.IBCDenom()
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, denomTrace)

	deposit := types.LSMTokenDeposit{
		ChainId:  HostChainId,
		Denom:    LSMTokenDenom,
		IbcDenom: ibcDenom,
		Status:   types.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
	}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

	callbackArgs := types.TransferLSMTokenCallback{
		Deposit: &deposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err, "no error expected when marshalling callback args")

	return callbackArgsBz
}

func (s *KeeperTestSuite) TestLSMTransferCallback_Successful() {
	callbackArgsBz := s.SetupLSMTransferCallback()

	// Call the callback with a successful response
	ackSuccess := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_SUCCESS,
	}
	err := s.App.RecordsKeeper.LSMTransferCallback(s.Ctx, channeltypes.Packet{}, ackSuccess, callbackArgsBz)
	s.Require().NoError(err, "no error expected when executing callback")

	// Confirm deposit has been updated to DETOKENIZATION_QUEUE
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit should have been found but was not")
	s.Require().Equal(types.LSMTokenDeposit_DETOKENIZATION_QUEUE.String(), record.Status.String(), "deposit status")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_InvalidCallbackArgs() {
	s.SetupLSMTransferCallback()

	// Call the callback with a successful ack, but invalid callback args
	invalidCallbackArgs := []byte{1, 2, 3}
	ackSuccess := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_SUCCESS,
	}
	err := s.App.RecordsKeeper.LSMTransferCallback(s.Ctx, channeltypes.Packet{}, ackSuccess, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal LSM transfer callback")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_AckTimeout() {
	callbackArgsBz := s.SetupLSMTransferCallback()

	// Call the callback with a timed-out response
	ackTimeout := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_TIMEOUT,
	}
	err := s.App.RecordsKeeper.LSMTransferCallback(s.Ctx, channeltypes.Packet{}, ackTimeout, callbackArgsBz)
	s.Require().NoError(err, "no error expected when executing callback")

	// Confirm deposit has been updated to status TRANSFER_QUEUE
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit should have been found but was not")
	s.Require().Equal(types.LSMTokenDeposit_TRANSFER_QUEUE.String(), record.Status.String(), "deposit status")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_AckFailed() {
	callbackArgsBz := s.SetupLSMTransferCallback()

	// Call the callback with an ack-failure response
	ackFailure := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_FAILURE,
	}
	err := s.App.RecordsKeeper.LSMTransferCallback(s.Ctx, channeltypes.Packet{}, ackFailure, callbackArgsBz)
	s.Require().NoError(err)

	// Confirm deposit has been updated to status FAILED
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit should have been found but was not")
	s.Require().Equal(types.LSMTokenDeposit_TRANSFER_FAILED.String(), record.Status.String(), "deposit status")
}
