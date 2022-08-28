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
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RedemptionCallbackState struct {
	epochUnbondingNumbers   []uint64
	userRedemptionRecordIds []string
	epochNumber             uint64
}

type RedemptionCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type RedemptionCallbackTestCase struct {
	initialState RedemptionCallbackState
	validArgs    RedemptionCallbackArgs
}

func (s *KeeperTestSuite) SetupRedemptionCallback() RedemptionCallbackTestCase {
	epochNumber := uint64(1)

	// userRedemptionRecords should NOT be claimable until after the callback is called
	recordId1 := recordtypes.UserRedemptionRecordKeyFormatter(chainId, epochNumber, "sender")
	userRedemptionRecord1 := recordtypes.UserRedemptionRecord{
		Id:          recordId1,
		IsClaimable: false,
	}
	recordId2 := recordtypes.UserRedemptionRecordKeyFormatter(chainId, epochNumber, "other_sender")
	userRedemptionRecord2 := recordtypes.UserRedemptionRecord{
		Id:          recordId2,
		IsClaimable: false,
	}

	// the hostZoneUnbonding should have HostZoneUnbonding_UNBONDED - meaning unbonding has completed, but the tokens
	// have not yet been transferred to the redemption account
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:            chainId,
		Status:                recordtypes.HostZoneUnbonding_UNBONDED,
		UserRedemptionRecords: []string{recordId1, recordId2},
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
	}
	hostZone := stakeibc.HostZone{
		ChainId:        chainId,
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), userRedemptionRecord1)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), userRedemptionRecord2)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{})
	ack := s.ICAPacketAcknowledgement(msgs)
	callbackArgs := types.RedemptionCallback{
		HostZoneId:              chainId,
		EpochUnbondingRecordIds: []uint64{epochNumber},
	}
	args, err := s.App.StakeibcKeeper.MarshalRedemptionCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return RedemptionCallbackTestCase{
		initialState: RedemptionCallbackState{
			epochUnbondingNumbers:   []uint64{epochNumber},
			userRedemptionRecordIds: []string{userRedemptionRecord1.Id, userRedemptionRecord2.Id},
			epochNumber:             epochNumber,
		},
		validArgs: RedemptionCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestRedemptionCallback_Successful() {
	tc := s.SetupRedemptionCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	for _, epochNumber := range initialState.epochUnbondingNumbers {
		// fetch the epoch unbonding record
		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), epochNumber)
		s.Require().True(found)
		for _, hzu := range epochUnbondingRecord.HostZoneUnbondings {
			// check that the status is transferred
			if hzu.HostZoneId == chainId {
				s.Require().Equal(recordtypes.HostZoneUnbonding_TRANSFERRED, hzu.Status)
			}
		}
	}

	// verify user redemption records are claimable
	for _, userRedemptionRecordId := range initialState.userRedemptionRecordIds {
		userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx(), userRedemptionRecordId)
		s.Require().True(found)
		s.Require().True(userRedemptionRecord.IsClaimable)
	}

}

func (s *KeeperTestSuite) checkRedemptionStateIfCallbackFailed(tc RedemptionCallbackTestCase) {
	initialState := tc.initialState
	for _, epochNumber := range initialState.epochUnbondingNumbers {
		// fetch the epoch unbonding record
		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), epochNumber)
		s.Require().True(found)
		for _, hzu := range epochUnbondingRecord.HostZoneUnbondings {
			// check that the status is NOT transferred
			s.Require().Equal(recordtypes.HostZoneUnbonding_UNBONDED, hzu.Status)
		}
	}

	// verify user redemption records are NOT claimable
	for _, userRedemptionRecordIds := range initialState.userRedemptionRecordIds {
		userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx(), userRedemptionRecordIds)
		s.Require().True(found)
		s.Require().False(userRedemptionRecord.IsClaimable)
	}
}

func (s *KeeperTestSuite) TestRedemptionCallback_RedemptionCallbackTimeout() {
	tc := s.SetupRedemptionCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil
	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_RedemptionCallbackErrorOnHost() {
	tc := s.SetupRedemptionCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck
	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_WrongCallbackArgs() {
	tc := s.SetupRedemptionCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "Unable to unmarshal redemption callback args | unexpected EOF: unable to unmarshal data structure")
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_EpochUnbondingRecordNotFound() {
	tc := s.SetupRedemptionCallback()
	invalidArgs := tc.validArgs
	callbackArgs := types.RedemptionCallback{
		HostZoneId:              chainId,
		EpochUnbondingRecordIds: []uint64{tc.initialState.epochNumber + 1},
	}
	args, err := s.App.StakeibcKeeper.MarshalRedemptionCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)
	invalidArgs.args = args
	err = stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	expectedErr := fmt.Sprintf("Epoch unbonding record not found for epoch #%d: key not found", tc.initialState.epochNumber+1)
	s.Require().EqualError(err, expectedErr)
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_HostZoneUnbondingNotFound() {
	tc := s.SetupRedemptionCallback()
	valid := tc.validArgs
	// remove the hzu from the epoch unbonding record
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), tc.initialState.epochNumber)
	s.Require().True(found)
	epochUnbondingRecord.HostZoneUnbondings = []*recordtypes.HostZoneUnbonding{}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), valid.packet, valid.ack, valid.args)
	s.Require().EqualError(err, fmt.Sprintf("Could not find host zone unbonding %d for host zone GAIA: not found", tc.initialState.epochNumber))
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_UserRedemptionRecordNotFound() {
	tc := s.SetupRedemptionCallback()
	valid := tc.validArgs
	recordId1 := tc.initialState.userRedemptionRecordIds[0]
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx(), recordId1)
	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), valid.packet, valid.ack, valid.args)
	s.Require().EqualError(err, fmt.Sprintf("no user redemption record found for id (%s): record not found", recordId1))
}
