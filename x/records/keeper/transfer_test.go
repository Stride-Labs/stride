package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"

	"github.com/Stride-Labs/stride/v4/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
)

type TransferTestCase struct {
	depositRecord types.DepositRecord
	transferMsg   ibctypes.MsgTransfer
}

func (s *KeeperTestSuite) SetupTransfer() TransferTestCase {
	s.CreateTransferChannel(chainId)
	balanceToTransfer := sdk.NewInt(1_000_000)
	depositRecord := types.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToTransfer,
		Status:             types.DepositRecord_TRANSFER_QUEUE,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)
	coin := sdk.NewCoin("tokens", balanceToTransfer)
	s.FundAccount(s.TestAccs[0], coin)
	transferMsg := ibctypes.MsgTransfer{
		SourcePort:    "transfer",
		SourceChannel: "channel-0",
		Token:         coin,
		Sender:        s.TestAccs[0].String(),
		Receiver:      s.TestAccs[1].String(),
		TimeoutHeight: clienttypes.NewHeight(0, 100),
	}

	return TransferTestCase{
		depositRecord: depositRecord,
		transferMsg:   transferMsg,
	}
}

func (s *KeeperTestSuite) TestTransfer_Successful() {
	tc := s.SetupTransfer()

	err := s.App.RecordsKeeper.Transfer(s.Ctx, &tc.transferMsg, tc.depositRecord)
	s.Require().NoError(err)

	// Confirm deposit record has been updated to TRANSFER_IN_PROGRESS
	record, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, tc.depositRecord.Id)
	s.Require().True(found)
	s.Require().Equal(record.Status, recordtypes.DepositRecord_TRANSFER_IN_PROGRESS, "deposit record status should be TRANSFER_IN_PROGRESS")
}

func (s *KeeperTestSuite) TestSequence_Equal() {
	tc := s.SetupTransfer()
	goCtx := sdk.WrapSDKContext(s.Ctx)
	sequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx,
		tc.transferMsg.SourcePort, tc.transferMsg.SourceChannel)
	s.Require().True(found)

	msgTransferResponse, err := s.App.TransferKeeper.Transfer(goCtx, &tc.transferMsg)
	s.Require().NoError(err)

	checkSequence := msgTransferResponse.Sequence

	// Confirm msg sequence are equal to next sequence
	s.Require().Equal(checkSequence, sequence, "sequence should be equal")
}
