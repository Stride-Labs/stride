package records_test

import (
	// "bytes"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"
	_ "github.com/stretchr/testify/suite"

	// "github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	// ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/records"
	"github.com/Stride-Labs/stride/v4/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
)

const chainId = "GAIA"

type TransferCallbackState struct {
	callbackArgs types.TransferCallback
}

type TransferCallbackArgs struct {
	packet channeltypes.Packet
	ack    []byte
}

type TransferCallbackTestCase struct {
	validArgs    TransferCallbackArgs
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

func (s *ModuleTestSuite) SetupTransferCallback() TransferCallbackTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", "GAIA", "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
  
	balanceToStake := sdk.NewInt(1_000_000)
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake.Int64(),
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)

	coins := sdk.NewCoin(sdk.DefaultBondDenom, balanceToStake)
	port := s.TransferPath.EndpointA.ChannelConfig.PortID
	channel := s.TransferPath.EndpointA.ChannelID
	accountFrom := s.StrideChain.SenderAccount.GetAddress().String()
	timeoutHeight := clienttypes.NewHeight(0, 100)

	msg := transfertypes.NewMsgTransfer(port, channel, coins, accountFrom, "INVALID", timeoutHeight, 0)
	err := s.App.RecordsKeeper.Transfer(s.Ctx, msg, depositRecord)
	s.Require().NoError(err)

	// Move forward one block
	s.StrideChain.NextBlock()
	s.StrideChain.SenderAccount.SetSequence(s.StrideChain.SenderAccount.GetSequence())
	s.StrideChain.Coordinator.IncrementTime()

	// Update both clients
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err)
	err = s.TransferPath.EndpointB.UpdateClient()
	s.Require().NoError(err)


	packetData := transfertypes.NewFungibleTokenPacketData(
		coins.Denom, coins.Amount.String(), msg.Sender, msg.Receiver,
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

	err = s.TransferPath.EndpointA.AcknowledgePacket(packet, ack)
	s.Require().NoError(err)

	return TransferCallbackTestCase{
		validArgs: TransferCallbackArgs{
			packet: packet,
			ack:    ack,
		},
	}
}

func (s *ModuleTestSuite) TestOnAcknowledgementPacket_Successful() {
	s.SetupTransferCallback()
	// err := s.IBCModule.OnAcknowledgementPacket(s.Ctx, tc.validArgs.packet, tc.validArgs.ack, nil)
	// s.Require().NoError(err)
}