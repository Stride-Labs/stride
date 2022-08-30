package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

type ClaimCallbackState struct {
	callbackArgs types.ClaimCallback
}

type ClaimCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type ClaimCallbackTestCase struct {
	initialState ClaimCallbackState
	validArgs    ClaimCallbackArgs
}

func (s *KeeperTestSuite) SetupClaimCallback() ClaimCallbackTestCase {
	epochNumber := uint64(1)
	recordId := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, "sender")
	userRedemptionRecord := recordtypes.UserRedemptionRecord{
		Id: recordId,
		// after a user calls ClaimUndelegatedTokens, the record is set to isClaimable = false
		// to prevent double claims
		IsClaimable: false,
	}
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), userRedemptionRecord)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{})
	ack := s.ICAPacketAcknowledgement(msgs, nil)
	callbackArgs := types.ClaimCallback{
		UserRedemptionRecordId: recordId,
	}
	args, err := s.App.StakeibcKeeper.MarshalClaimCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return ClaimCallbackTestCase{
		initialState: ClaimCallbackState{
			callbackArgs: callbackArgs,
		},
		validArgs: ClaimCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestClaimCallback_Successful() {
	tc := s.SetupClaimCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	_, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx(), initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().False(found, "record has been deleted")
}

func (s *KeeperTestSuite) checkClaimStateIfCallbackFailed(tc ClaimCallbackTestCase) {
	record, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx(), tc.initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().True(found)
	s.Require().True(record.IsClaimable, "record is set to isClaimable = true (if the callback failed, it should be reset to true so that users can retry the claim)")
}

func (s *KeeperTestSuite) TestClaimCallback_ClaimCallbackTimeout() {
	tc := s.SetupClaimCallback()
	invalidArgs := tc.validArgs
	invalidArgs.ack = nil
	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err, "timeout successfully proccessed")
	s.checkClaimStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestClaimCallback_ClaimCallbackErrorOnHost() {
	tc := s.SetupClaimCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err, "error ack successfully proccessed")
	s.checkClaimStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestClaimCallback_WrongCallbackArgs() {
	tc := s.SetupClaimCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "unexpected EOF")
}

func (s *KeeperTestSuite) TestClaimCallback_RecordNotFound() {
	tc := s.SetupClaimCallback()
	validArgs := tc.validArgs
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx(), tc.initialState.callbackArgs.UserRedemptionRecordId)
	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().EqualError(err, fmt.Sprintf("user redemption record not found %s: record not found", tc.initialState.callbackArgs.UserRedemptionRecordId))
}
