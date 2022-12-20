package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordskeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
	"github.com/Stride-Labs/stride/v4/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
)

const chainId = "GAIA"

type TransferCallbackState struct {
	callbackArgs types.TransferCallback
}

type TransferCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type TransferCallbackTestCase struct {
	initialState TransferCallbackState
	validArgs    TransferCallbackArgs
}

func (s *KeeperTestSuite) SetupTransferCallback() TransferCallbackTestCase {
	balanceToStake := sdk.NewInt(1_000_000)
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake,
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	packet := channeltypes.Packet{Data: s.MarshalledICS20PacketData()}
	ack := s.ICS20PacketAcknowledgement()
	callbackArgs := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	args, err := s.App.RecordsKeeper.MarshalTransferCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return TransferCallbackTestCase{
		initialState: TransferCallbackState{
			callbackArgs: callbackArgs,
		},
		validArgs: TransferCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestTransferCallback_Successful() {
	tc := s.SetupTransferCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, validArgs.packet, validArgs.ack, validArgs.args)
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
	timeoutArgs := tc.validArgs
	// a nil ack means the request timed out
	timeoutArgs.ack = nil
	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, timeoutArgs.packet, timeoutArgs.ack, timeoutArgs.args)
	s.Require().NoError(err)
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_TransferCallbackErrorOnHost() {
	tc := s.SetupTransferCallback()
	errorArgs := tc.validArgs
	// an error ack means the tx failed on the host
	errorAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, errorArgs.packet, &errorAck, errorArgs.args)
	s.Require().NoError(err)
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, types.DepositRecord_TRANSFER_QUEUE, "DepositRecord is put back in the TRANSFER_QUEUE after a failed transfer")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_WrongCallbackArgs() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "cannot unmarshal transfer callback args: unexpected EOF: cannot unmarshal")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_DepositRecordNotFound() {
	tc := s.SetupTransferCallback()
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, tc.initialState.callbackArgs.DepositRecordId)

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ack, tc.validArgs.args)
	s.Require().EqualError(err, fmt.Sprintf("deposit record not found %d: unknown deposit record", tc.initialState.callbackArgs.DepositRecordId))
}

func (s *KeeperTestSuite) TestTransferCallback_PacketUnmarshallingError() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs
	invalidArgs.packet.Data = []byte("random bytes")

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().EqualError(err, "cannot unmarshal ICS-20 transfer packet data: invalid character 'r' looking for beginning of value: unknown request")
}
