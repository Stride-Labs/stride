package records_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"
	_ "github.com/stretchr/testify/suite"

	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/records"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
)

type TransferCallbackTestCase struct {
	packet channeltypes.Packet
	ack    []byte
}

type ModuleTestSuite struct {
	apptesting.AppTestHelper
	IBCModule records.IBCModule
}

func (s *ModuleTestSuite) SetupTest() {
	s.Setup()
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (s *ModuleTestSuite) SetupTransferMsg() (transfertypes.MsgTransfer, recordtypes.DepositRecord) {
	delegationAccountOwner := fmt.Sprintf("%s.%s", "GAIA", "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	balanceToStake := sdk.NewInt(1_000_000)
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             balanceToStake.Int64(),
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord)

	coins := sdk.NewCoin(sdk.DefaultBondDenom, balanceToStake)
	port := s.TransferPath.EndpointA.ChannelConfig.PortID
	channel := s.TransferPath.EndpointA.ChannelID
	accountFrom := s.StrideChain.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(0, 100)

	msg := transfertypes.NewMsgTransfer(port, channel, coins, accountFrom, s.IcaAddresses[delegationAccountOwner], timeoutHeight, 0)

	return *msg, depositRecord
}

func (s *ModuleTestSuite) GetPacketAndAck(msg transfertypes.MsgTransfer, depositRecord recordtypes.DepositRecord) TransferCallbackTestCase {
	err := s.App.RecordsKeeper.Transfer(s.TransferPath.EndpointA.Chain.GetContext(), &msg, depositRecord)
	s.Require().NoError(err)

	// Update both clients
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err)
	err = s.TransferPath.EndpointB.UpdateClient()
	s.Require().NoError(err)

	packetData := transfertypes.NewFungibleTokenPacketData(
		msg.Token.Denom, msg.Token.Amount.String(), msg.Sender, msg.Receiver,
	)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		msg.SourcePort,
		msg.SourceChannel,
		s.TransferPath.EndpointB.Counterparty.ChannelConfig.PortID,
		s.TransferPath.EndpointB.Counterparty.ChannelID,
		msg.TimeoutHeight,
		msg.TimeoutTimestamp,
	)

	// // get the ack from the chain b's response
	res, err := s.TransferPath.EndpointB.RecvPacketWithResult(packet)
	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	s.Require().NoError(err)

	return TransferCallbackTestCase{
		packet: packet,
		ack:    ack,
	}
}

func (s *ModuleTestSuite) TestOnAcknowledgementPacket_Successful() {
	msg, depositRecord := s.SetupTransferMsg()
	tc := s.GetPacketAndAck(msg, depositRecord)
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	err := s.TransferPath.EndpointA.AcknowledgePacket(tc.packet, tc.ack)
	s.Require().NoError(err)

	// check record after refund
	record, found = s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_DELEGATION_QUEUE)
}

func (s *ModuleTestSuite) TestOnAcknowledgementPacket_AckErr() {
	msg, depositRecord := s.SetupTransferMsg()
	msg.Receiver = "INVALID"
	tc := s.GetPacketAndAck(msg, depositRecord)
	balanceBefore := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	err := s.TransferPath.EndpointA.AcknowledgePacket(tc.packet, tc.ack)
	s.Require().NoError(err)

	// check record after refund
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_TRANSFER_QUEUE)

	// check balance
	balanceAfter := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	s.Require().Equal(balanceBefore.Add(msg.Token), balanceAfter)
}

func (s *ModuleTestSuite) TestOnAcknowledgementPacket_ErrCallBack() {
	msg, depositRecord := s.SetupTransferMsg()
	tc := s.GetPacketAndAck(msg, depositRecord)

	balanceBefore := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	s.App.RecordsKeeper.RemoveDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)

	err := s.TransferPath.EndpointA.AcknowledgePacket(tc.packet, tc.ack)
	s.Require().NoError(err)

	// check record after refund
	_, found := s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	s.Require().False(found)

	// check balance
	balanceAfter := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	s.Require().Equal(balanceBefore.Add(msg.Token), balanceAfter)
}

func (s *ModuleTestSuite) GetTimeOutPacket(msg transfertypes.MsgTransfer, depositRecord recordtypes.DepositRecord) TransferCallbackTestCase {
	err := s.App.RecordsKeeper.Transfer(s.TransferPath.EndpointA.Chain.GetContext(), &msg, depositRecord)
	s.Require().NoError(err)

	// Update both clients
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err)
	err = s.TransferPath.EndpointB.UpdateClient()
	s.Require().NoError(err)

	packetData := transfertypes.NewFungibleTokenPacketData(
		msg.Token.Denom, msg.Token.Amount.String(), msg.Sender, msg.Receiver,
	)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		msg.SourcePort,
		msg.SourceChannel,
		s.TransferPath.EndpointB.Counterparty.ChannelConfig.PortID,
		s.TransferPath.EndpointB.Counterparty.ChannelID,
		clienttypes.GetSelfHeight(s.TransferPath.EndpointB.Chain.GetContext()),
		msg.TimeoutTimestamp,
	)

	s.TransferPath.EndpointA.SendPacket(packet)
	s.TransferPath.EndpointA.UpdateClient()

	err = s.TransferPath.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)

	return TransferCallbackTestCase{
		packet: packet,
		ack:    nil,
	}
}

func (s *ModuleTestSuite) TestOnTimeoutPacket() {
	msg, depositRecord := s.SetupTransferMsg()
	s.IBCModule = records.NewIBCModule(s.App.RecordsKeeper, transfer.NewIBCModule(s.App.TransferKeeper))
	tc := s.GetPacketAndAck(msg, depositRecord)
	balanceBefore := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))

	err := s.IBCModule.OnTimeoutPacket(s.TransferPath.EndpointA.Chain.GetContext(), tc.packet, nil)
	s.Require().NoError(err)

	// check balance
	balanceAfter := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	s.Require().Equal(balanceBefore.Add(msg.Token), balanceAfter)

	// check record after refund
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_TRANSFER_QUEUE)
}

func (s *ModuleTestSuite) TestOnTimeoutPacket_RecordErr() {
	msg, depositRecord := s.SetupTransferMsg()
	s.IBCModule = records.NewIBCModule(s.App.RecordsKeeper, transfer.NewIBCModule(s.App.TransferKeeper))
	tc := s.GetPacketAndAck(msg, depositRecord)
	s.App.RecordsKeeper.RemoveDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	balanceBefore := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))

	err := s.IBCModule.OnTimeoutPacket(s.TransferPath.EndpointA.Chain.GetContext(), tc.packet, nil)
	s.Require().NoError(err)

	// check balance
	balanceAfter := s.App.BankKeeper.SpendableCoins(s.TransferPath.EndpointA.Chain.GetContext(), sdk.MustAccAddressFromBech32(msg.Sender))
	s.Require().Equal(balanceBefore.Add(msg.Token), balanceAfter)

	// check record after refund
	_, found := s.App.RecordsKeeper.GetDepositRecord(s.TransferPath.EndpointA.Chain.GetContext(), depositRecord.Id)
	s.Require().False(found)
}
