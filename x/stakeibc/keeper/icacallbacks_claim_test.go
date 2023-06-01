package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type ClaimCallbackState struct {
	callbackArgs    types.ClaimCallback
	epochNumber     uint64
	decrementAmount sdkmath.Int
	hzu1TokenAmount sdkmath.Int
}

type ClaimCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type ClaimCallbackTestCase struct {
	initialState ClaimCallbackState
	validArgs    ClaimCallbackArgs
}

func (s *KeeperTestSuite) SetupClaimCallback() ClaimCallbackTestCase {
	epochNumber := uint64(1)
	recordId1 := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, "sender")
	userRedemptionRecord1 := recordtypes.UserRedemptionRecord{
		Id: recordId1,
		// after a user calls ClaimUndelegatedTokens, the record is set to claimIsPending = true
		// to prevent double claims
		ClaimIsPending: true,
		Amount:         sdkmath.ZeroInt(),
	}
	recordId2 := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, "other_sender")
	userRedemptionRecord2 := recordtypes.UserRedemptionRecord{
		Id:             recordId2,
		ClaimIsPending: false,
	}
	recordId3 := recordtypes.UserRedemptionRecordKeyFormatter("not_gaia", epochNumber, "sender")
	userRedemptionRecord3 := recordtypes.UserRedemptionRecord{
		Id:             recordId3,
		ClaimIsPending: false,
	}
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, userRedemptionRecord1)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, userRedemptionRecord2)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, userRedemptionRecord3)
	// the hzu that we'll claim from
	hostZoneUnbonding1 := recordtypes.HostZoneUnbonding{
		HostZoneId:            HostChainId,
		Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
		UserRedemptionRecords: []string{recordId1, recordId2},
		NativeTokenAmount:     sdkmath.NewInt(1_000_000),
	}
	hostZoneUnbonding2 := recordtypes.HostZoneUnbonding{
		HostZoneId:            "not_gaia",
		Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		UserRedemptionRecords: []string{recordId3},
		NativeTokenAmount:     sdkmath.NewInt(1_000_000),
	}
	// some other hzus in the future
	hostZoneUnbonding3 := recordtypes.HostZoneUnbonding{
		HostZoneId:        "not_gaia",
		Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		NativeTokenAmount: sdkmath.NewInt(1_000_000),
	}
	hostZoneUnbonding4 := recordtypes.HostZoneUnbonding{
		HostZoneId:        HostChainId,
		Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		NativeTokenAmount: sdkmath.NewInt(1_000_000),
	}
	epochUnbondingRecord1 := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding1, &hostZoneUnbonding2},
	}
	epochUnbondingRecord2 := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber + 1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding3, &hostZoneUnbonding4},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord1)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord2)

	// Mock ack response and callback args
	packet := channeltypes.Packet{}
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_SUCCESS}
	callbackArgs := types.ClaimCallback{
		UserRedemptionRecordId: recordId1,
		ChainId:                HostChainId,
		EpochNumber:            epochNumber,
	}
	callbackArgsBz, err := s.App.StakeibcKeeper.MarshalClaimCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	decrementAmount := userRedemptionRecord1.Amount

	return ClaimCallbackTestCase{
		initialState: ClaimCallbackState{
			callbackArgs:    callbackArgs,
			epochNumber:     epochNumber,
			decrementAmount: decrementAmount,
			hzu1TokenAmount: hostZoneUnbonding1.NativeTokenAmount,
		},
		validArgs: ClaimCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        callbackArgsBz,
		},
	}
}

func (s *KeeperTestSuite) TestClaimCallback_Successful() {
	tc := s.SetupClaimCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err)

	_, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().False(found, "record has been deleted")

	// fetch the epoch unbonding record
	epochUnbondingRecord1, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	epochUnbondingRecord2, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber+1)
	s.Require().True(found, "epoch unbonding record found")

	// fetch the hzus
	hzu1 := epochUnbondingRecord1.HostZoneUnbondings[0]
	hzu2 := epochUnbondingRecord1.HostZoneUnbondings[1]
	hzu3 := epochUnbondingRecord2.HostZoneUnbondings[0]
	hzu4 := epochUnbondingRecord2.HostZoneUnbondings[1]

	// check that hzu1 has a decremented amount
	s.Require().Equal(hzu1.NativeTokenAmount, tc.initialState.hzu1TokenAmount.Sub(tc.initialState.decrementAmount), "hzu1 amount decremented")
	s.Require().Equal(hzu1.Status, recordtypes.HostZoneUnbonding_CLAIMABLE, "hzu1 status set to transferred")
	// verify the other hzus are unchanged
	s.Require().Equal(hzu2.NativeTokenAmount, hzu2.NativeTokenAmount, "hzu2 amount unchanged")
	s.Require().Equal(hzu2.Status, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, "hzu2 status set to transferred")
	s.Require().Equal(hzu3.NativeTokenAmount, hzu3.NativeTokenAmount, "hzu3 amount unchanged")
	s.Require().Equal(hzu3.Status, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, "hzu3 status set to transferred")
	s.Require().Equal(hzu4.NativeTokenAmount, hzu4.NativeTokenAmount, "hzu4 amount unchanged")
	s.Require().Equal(hzu4.Status, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, "hzu4 status set to transferred")
}

func (s *KeeperTestSuite) checkClaimStateIfCallbackFailed(tc ClaimCallbackTestCase) {
	record, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, tc.initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().True(found)
	s.Require().False(record.ClaimIsPending, "record is set to claimIsPending = false (if the callback failed, it should be reset to false so that users can retry the claim)")
}

func (s *KeeperTestSuite) TestClaimCallback_ClaimCallbackTimeout() {
	tc := s.SetupClaimCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "timeout successfully proccessed")
	s.checkClaimStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestClaimCallback_ClaimCallbackErrorOnHost() {
	tc := s.SetupClaimCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "error ack successfully proccessed")
	s.checkClaimStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestClaimCallback_WrongCallbackArgs() {
	tc := s.SetupClaimCallback()

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal claim callback args: unexpected EOF: unable to unmarshal data structure")
}

func (s *KeeperTestSuite) TestClaimCallback_RecordNotFound() {
	tc := s.SetupClaimCallback()

	// Remove the user redemption record from the state
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, tc.initialState.callbackArgs.UserRedemptionRecordId)

	err := stakeibckeeper.ClaimCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, fmt.Sprintf("user redemption record not found %s: record not found", tc.initialState.callbackArgs.UserRedemptionRecordId))
}

// DecrementHostZoneUnbonding decreases the number of tokens claimed by a user on a particular hzu
func (s *KeeperTestSuite) TestDecrementHostZoneUnbonding_Success() {
	tc := s.SetupClaimCallback()
	initialState := tc.initialState

	userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().True(found, "record has been deleted")

	err := s.App.StakeibcKeeper.DecrementHostZoneUnbonding(s.Ctx, userRedemptionRecord, tc.initialState.callbackArgs)
	s.Require().NoError(err, "host zone unbonding successfully decremented")

	// fetch the epoch unbonding record
	epochUnbondingRecord1, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")

	// fetch the hzus
	hzu1 := epochUnbondingRecord1.HostZoneUnbondings[0]

	// check that hzu1 has a decremented amount
	s.Require().Equal(hzu1.NativeTokenAmount.Sub(userRedemptionRecord.Amount), hzu1.NativeTokenAmount, "hzu1 amount decremented")
}

func (s *KeeperTestSuite) TestDecrementHostZoneUnbonding_HzuNotFound() {
	tc := s.SetupClaimCallback()
	initialState := tc.initialState

	// remove the hzus
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	epochUnbondingRecord.HostZoneUnbondings = []*recordtypes.HostZoneUnbonding{}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, initialState.callbackArgs.UserRedemptionRecordId)
	s.Require().True(found, "record has been deleted")

	err := s.App.StakeibcKeeper.DecrementHostZoneUnbonding(s.Ctx, userRedemptionRecord, tc.initialState.callbackArgs)
	s.Require().EqualError(err, "host zone unbonding not found GAIA: record not found")
}
