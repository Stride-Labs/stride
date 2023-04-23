package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	"github.com/golang/protobuf/proto" //nolint:staticcheck

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordskeeper "github.com/Stride-Labs/stride/v9/x/records/keeper"
	"github.com/Stride-Labs/stride/v9/x/records/types"
)

var (
	LSMTokenDenom = "cosmosvaloperxxx/42"
)

func (s *KeeperTestSuite) SetupLSMTransferCallback() []byte {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

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
		Deposit:                      &deposit,
		TransferChannelId:            ibctesting.FirstChannelID,
		HostZoneDepositAddress:       s.TestAccs[0].String(),
		HostZoneDelegationIcaAddress: delegationAddress,
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
	err := recordskeeper.LSMTransferCallback(s.App.RecordsKeeper, s.Ctx, channeltypes.Packet{}, ackSuccess, callbackArgsBz)
	s.Require().NoError(err, "no error expected when executing callback")

	// Confirm deposit has been updated to DETOKENIZATION_QUEUE
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit found")
	s.Require().Equal(types.LSMTokenDeposit_DETOKENIZATION_QUEUE.String(), record.Status.String(), "deposit status")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_InvalidCallbackArgs() {
	s.SetupLSMTransferCallback()

	// Call the callback with a successful ack, but invalid callback args
	invalidCallbackArgs := []byte{1, 2, 3}
	ackSuccess := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_SUCCESS,
	}
	err := recordskeeper.LSMTransferCallback(s.App.RecordsKeeper, s.Ctx, channeltypes.Packet{}, ackSuccess, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal LSM transfer callback")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_AckTimeout() {
	callbackArgsBz := s.SetupLSMTransferCallback()

	// Call the callback with a timed-out response
	ackTimeout := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_TIMEOUT,
	}
	err := recordskeeper.LSMTransferCallback(s.App.RecordsKeeper, s.Ctx, channeltypes.Packet{}, ackTimeout, callbackArgsBz)
	s.Require().NoError(err, "no error expected when executing callback")

	// Confirm deposit has been updated to TRANSFER_IN_PROGRESS (since the transfer gets resubmitted on a timeout)
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit found")
	s.Require().Equal(types.DepositRecord_TRANSFER_IN_PROGRESS.String(), record.Status.String(), "deposit status")
}

func (s *KeeperTestSuite) TestLSMTransferCallback_AckFailed() {
	callbackArgsBz := s.SetupLSMTransferCallback()

	// Call the callback with an ack-failure response
	ackFailure := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_FAILURE,
	}
	err := recordskeeper.LSMTransferCallback(s.App.RecordsKeeper, s.Ctx, channeltypes.Packet{}, ackFailure, callbackArgsBz)
	s.Require().NoError(err)

	// Confirm deposit has been updated to status FAILED
	record, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenDenom)
	s.Require().True(found, "deposit found")
	s.Require().Equal(types.LSMTokenDeposit_TRANSFER_FAILED.String(), record.Status.String(), "deposit status")
}
