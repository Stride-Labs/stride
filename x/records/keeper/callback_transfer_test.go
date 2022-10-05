package keeper_test

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordskeeper "github.com/Stride-Labs/stride/x/records/keeper"
	"github.com/Stride-Labs/stride/x/records/types"
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
	balanceToStake := int64(1_000_000)
	depositRecord := types.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake,
		Status:             types.DepositRecord_TRANSFER,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx(), depositRecord)
	packet := channeltypes.Packet{Data: s.MarshalledICS20PacketData()}
	ack := s.ICS20PacketAcknowledgement()
	callbackArgs := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	args, err := s.App.RecordsKeeper.MarshalTransferCallbackArgs(s.Ctx(), callbackArgs)
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

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Confirm deposit record has been updated to STAKE
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, types.DepositRecord_STAKE, "deposit record status should be STAKE")
}

func (s *KeeperTestSuite) checkTransferStateIfCallbackFailed(tc TransferCallbackTestCase) {
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), tc.initialState.callbackArgs.DepositRecordId)
	s.Require().True(found)
	s.Require().Equal(record.Status, types.DepositRecord_TRANSFER, "deposit record status should be TRANSFER")
}

func (s *KeeperTestSuite) TestTransferCallback_TransferCallbackTimeout() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil
	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_TransferCallbackErrorOnHost() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	errorAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), invalidArgs.packet, &errorAck, invalidArgs.args)
	s.Require().EqualError(err, "TransferCallback does not handle errors: error: invalid request")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_WrongCallbackArgs() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "cannot unmarshal transfer callback args: unexpected EOF: cannot unmarshal")
	s.checkTransferStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestTransferCallback_DepositRecordNotFound() {
	tc := s.SetupTransferCallback()
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx(), tc.initialState.callbackArgs.DepositRecordId)

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), tc.validArgs.packet, tc.validArgs.ack, tc.validArgs.args)
	s.Require().EqualError(err, fmt.Sprintf("deposit record not found %d: unknown deposit record", tc.initialState.callbackArgs.DepositRecordId))
}

func (s *KeeperTestSuite) TestTransferCallback_PacketUnmarshallingError() {
	tc := s.SetupTransferCallback()
	invalidArgs := tc.validArgs
	invalidArgs.packet.Data = []byte("random bytes")

	err := recordskeeper.TransferCallback(s.App.RecordsKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().EqualError(err, "cannot unmarshal ICS-20 transfer packet data: invalid character 'r' looking for beginning of value: unknown request")
}
