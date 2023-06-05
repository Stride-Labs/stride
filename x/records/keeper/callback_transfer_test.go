package keeper_test

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordskeeper "github.com/Stride-Labs/stride/v9/x/records/keeper"
	"github.com/Stride-Labs/stride/v9/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
)

const chainId = "GAIA"

type TransferCallbackState struct {
	callbackArgs types.TransferCallback
}

type TransferCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type TransferCallbackTestCase struct {
	initialState TransferCallbackState
	validArgs    TransferCallbackArgs
}

func (s *KeeperTestSuite) SetupTransferCallback() TransferCallbackTestCase {
	balanceToStake := sdkmath.NewInt(1_000_000)
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake,
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	packet := channeltypes.Packet{Data: s.MarshalledICS20PacketData()}
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_SUCCESS}
	callbackArgs := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	callbackArgsBz, err := s.App.RecordsKeeper.MarshalTransferCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return TransferCallbackTestCase{
		initialState: TransferCallbackState{
			callbackArgs: callbackArgs,
		},
		validArgs: TransferCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        callbackArgsBz,
		},
	}
}

func (s *KeeperTestSuite) TestTransferCallback_Successful() {
	tc := s.SetupTransferCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err)

	// Confirm deposit record has been updated to DELEGATION_QUEUE
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_DELEGATION_QUEUE, "deposit record status should be DELEGATION_QUEUE")
}

func (s *KeeperTestSuite) checkTransferStateIfCallbackFailed(tc TransferCallbackTestCase) {
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_TRANSFER_QUEUE, "deposit record status should be TRANSFER_QUEUE")
}

func (s *KeeperTestSuite) TestTransferCallback_TransferCallbackTimeout() {
	tc := s.SetupTransferCallback()

	// Update the ack response to indicate a timeout
	timeoutArgs := tc.validArgs
	timeoutArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, timeoutArgs.packet, timeoutArgs.ackResponse, timeoutArgs.args)
	s.Require().NoError(err)
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_TransferCallbackErrorOnHost() {
	tc := s.SetupTransferCallback()

	// an error ack means the tx failed on the host
	errorArgs := tc.validArgs
	errorArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, errorArgs.packet, errorArgs.ackResponse, errorArgs.args)
	s.Require().NoError(err)

	// Confirm deposit record status is reverted
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, types.DepositRecord_TRANSFER_QUEUE, "DepositRecord is put back in the TRANSFER_QUEUE after a failed transfer")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_WrongCallbackArgs() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "cannot unmarshal transfer callback args: unexpected EOF: cannot unmarshal")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_DepositRecordNotFound() {
	tc := s.SetupTransferCallback()

	// Remove deposit record from store
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, tc.initialState.callbackArgs.DepositRecordId)

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, fmt.Sprintf("deposit record not found %d: unknown deposit record", tc.initialState.callbackArgs.DepositRecordId))
}

func (s *KeeperTestSuite) TestTransferCallback_PacketUnmarshallingError() {
	tc := s.SetupTransferCallback()

	// Update the data field within the packet so that the ICS transfer packet cannot be unmarshalled
	invalidArgs := tc.validArgs
	invalidArgs.packet.Data = []byte("random bytes")

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().EqualError(err, "cannot unmarshal ICS-20 transfer packet data: invalid character 'r' looking for beginning of value: unknown request")
}
