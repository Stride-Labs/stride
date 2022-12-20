package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type ReinvestCallbackState struct {
	reinvestAmt   sdk.Int
	callbackArgs  types.ReinvestCallback
	depositRecord recordtypes.DepositRecord
}

type ReinvestCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type ReinvestCallbackTestCase struct {
	initialState ReinvestCallbackState
	validArgs    ReinvestCallbackArgs
}

func (s *KeeperTestSuite) SetupReinvestCallback() ReinvestCallbackTestCase {
	reinvestAmt := sdk.NewInt(1_000)

	hostZone := stakeibc.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
	}
	expectedNewDepositRecord := recordtypes.DepositRecord{
		Id:                 0,
		DepositEpochNumber: 1,
		HostZoneId:         HostChainId,
		Amount:             reinvestAmt,
		Status:             recordtypes.DepositRecord_DELEGATION_QUEUE,
		Source:             recordtypes.DepositRecord_WITHDRAWAL_ICA,
	}
	epochTracker := stakeibc.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{}, &banktypes.MsgSend{})
	ack := s.ICAPacketAcknowledgement(msgs, nil)
	callbackArgs := types.ReinvestCallback{
		HostZoneId:     HostChainId,
		ReinvestAmount: sdk.NewCoin(Atom, reinvestAmt),
	}
	args, err := s.App.StakeibcKeeper.MarshalReinvestCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return ReinvestCallbackTestCase{
		initialState: ReinvestCallbackState{
			reinvestAmt:   reinvestAmt,
			callbackArgs:  callbackArgs,
			depositRecord: expectedNewDepositRecord,
		},
		validArgs: ReinvestCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestReinvestCallback_Successful() {
	tc := s.SetupReinvestCallback()
	initialState := tc.initialState
	expectedRecord := initialState.depositRecord
	validArgs := tc.validArgs

	err := stakeibckeeper.ReinvestCallback(s.App.StakeibcKeeper, s.Ctx, validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Confirm deposit record has been added
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]

	// Confirm deposit record fields match those expected
	s.Require().Equal(int64(expectedRecord.Id), int64(record.Id), "deposit record Id")
	s.Require().Equal(expectedRecord.Amount, record.Amount, "deposit record Amount")
	s.Require().Equal(expectedRecord.HostZoneId, record.HostZoneId, "deposit record HostZoneId")
	s.Require().Equal(expectedRecord.Status, record.Status, "deposit record Status")
	s.Require().Equal(expectedRecord.Source, record.Source, "deposit record Source")
	s.Require().Equal(int64(expectedRecord.DepositEpochNumber), int64(record.DepositEpochNumber), "deposit record DepositEpochNumber")
}

func (s *KeeperTestSuite) checkReinvestStateIfCallbackFailed(tc ReinvestCallbackTestCase) {
	// Confirm deposit record has not been added
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 0, "number of deposit records")
}

func (s *KeeperTestSuite) TestReinvestCallback_ReinvestCallbackTimeout() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil

	err := stakeibckeeper.ReinvestCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkReinvestStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestReinvestCallback_ReinvestCallbackErrorOnHost() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck

	err := stakeibckeeper.ReinvestCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkReinvestStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestReinvestCallback_WrongCallbackArgs() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.ReinvestCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "unexpected EOF")
	s.checkReinvestStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestReinvestCallback_MissingEpoch() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs

	// Remove epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := stakeibckeeper.ReinvestCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().ErrorContains(err, "no number for epoch (stride_epoch)")
	s.checkReinvestStateIfCallbackFailed(tc)
}
