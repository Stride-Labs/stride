package keeper_test

import (
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	proto "github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v30/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v30/x/records/types"
	"github.com/Stride-Labs/stride/v30/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v30/x/stakeibc/types"
)

type ClaimUndelegatedState struct {
	hostZone           types.HostZone
	redemptionRecordId string
	redemptionRecord   recordtypes.UserRedemptionRecord
}

type ClaimUndelegatedTestCase struct {
	validMsg       types.MsgClaimUndelegatedTokens
	initialState   ClaimUndelegatedState
	expectedIcaMsg keeper.IcaTx
}

func (s *KeeperTestSuite) SetupClaimUndelegatedTokens() ClaimUndelegatedTestCase {
	redemptionIcaOwner := "GAIA.REDEMPTION"
	s.CreateICAChannel(redemptionIcaOwner)

	epochNumber := uint64(1)
	senderAddr := "stride_SENDER"
	receiverAddr := "cosmos_RECEIVER"
	redemptionAddr := s.IcaAddresses[redemptionIcaOwner]
	redemptionRecordId := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, receiverAddr)

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		RedemptionIcaAddress: redemptionAddr,
		ConnectionId:         ibctesting.FirstConnectionID,
	}

	redemptionRecord := recordtypes.UserRedemptionRecord{
		Id:                redemptionRecordId,
		HostZoneId:        HostChainId,
		EpochNumber:       epochNumber,
		Receiver:          receiverAddr,
		Denom:             "uatom",
		ClaimIsPending:    false,
		NativeTokenAmount: sdkmath.NewInt(1000),
	}
	redemptionAmount := sdk.NewCoins(sdk.NewCoin(redemptionRecord.Denom, sdkmath.NewInt(1000)))

	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        epochNumber,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	hostZoneUnbonding1 := recordtypes.HostZoneUnbonding{
		HostZoneId:            HostChainId,
		Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
		UserRedemptionRecords: []string{redemptionRecordId},
		NativeTokenAmount:     sdkmath.NewInt(1_000_000),
	}
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding1},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, redemptionRecord)

	return ClaimUndelegatedTestCase{
		validMsg: types.MsgClaimUndelegatedTokens{
			Creator:    senderAddr,
			HostZoneId: HostChainId,
			Epoch:      epochNumber,
			Receiver:   receiverAddr,
		},
		initialState: ClaimUndelegatedState{
			hostZone:           hostZone,
			redemptionRecordId: redemptionRecordId,
			redemptionRecord:   redemptionRecord,
		},
		expectedIcaMsg: keeper.IcaTx{
			Msgs: []proto.Message{&banktypes.MsgSend{
				FromAddress: redemptionAddr,
				ToAddress:   receiverAddr,
				Amount:      redemptionAmount,
			}},
			ICAAccountType: types.ICAAccountType_REDEMPTION,
			Timeout:        uint64(types.DefaultICATimeoutNanos),
		},
	}
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_Successful() {
	tc := s.SetupClaimUndelegatedTokens()
	redemptionRecordId := tc.initialState.redemptionRecordId
	expectedRedemptionRecord := tc.initialState.redemptionRecord

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "claim undelegated tokens")

	actualRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, redemptionRecordId)
	s.Require().True(found, "redemption record found")
	s.Require().True(actualRedemptionRecord.ClaimIsPending, "redemption record should be pending")
	s.Require().Equal(expectedRedemptionRecord.NativeTokenAmount, actualRedemptionRecord.NativeTokenAmount, "record has expected amount")
	// TODO: check callback data here
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_SuccessfulMsgSendICA() {
	tc := s.SetupClaimUndelegatedTokens()
	redemptionRecord := tc.initialState.redemptionRecord

	icaTx, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &redemptionRecord, redemptionRecord.HostZoneId)
	msgs := icaTx.Msgs
	s.Require().NoError(err, "get redemption transfer msgs error")
	s.Require().Equal(1, len(msgs), "number of transfer messages")
	s.Require().Equal(tc.expectedIcaMsg.Msgs, msgs, "transfer message")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoUserRedemptionRecord() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove the user redemption record
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, tc.initialState.redemptionRecordId)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &tc.validMsg)
	s.Require().ErrorContains(err, "unable to find claimable redemption record")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_RecordNotClaimable() {
	tc := s.SetupClaimUndelegatedTokens()
	// Mark redemption record as not claimable
	alreadyClaimedRedemptionRecord := tc.initialState.redemptionRecord
	alreadyClaimedRedemptionRecord.ClaimIsPending = true
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, alreadyClaimedRedemptionRecord)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &tc.validMsg)
	s.Require().ErrorContains(err, "user redemption record GAIA.1.cosmos_RECEIVER is not claimable")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_RecordNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	invalidMsg := tc.validMsg
	invalidMsg.HostZoneId = "fake_host_zone"

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "user redemption record fake_host_zone.1.cosmos_RECEIVER not found on host zone fake_host_zone")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_HostZoneNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	invalidMsg := tc.validMsg
	invalidMsg.HostZoneId = "fake_host_zone"

	badRedemptionRecordId := strings.Replace(tc.initialState.redemptionRecordId, "GAIA", "fake_host_zone", 1)
	badRedemptionRecord := tc.initialState.redemptionRecord
	badRedemptionRecord.Id = badRedemptionRecordId
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, badRedemptionRecord)

	_, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &badRedemptionRecord, invalidMsg.HostZoneId)
	s.Require().EqualError(err, "Host zone fake_host_zone not found: host zone not registered")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoRedemptionAccount() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove redemption account from host zone
	hostZone := tc.initialState.hostZone
	hostZone.RedemptionIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &tc.initialState.redemptionRecord, tc.validMsg.HostZoneId)
	s.Require().ErrorContains(err, "Redemption account not found for host zone GAIA")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoEpochTracker() {
	tc := s.SetupClaimUndelegatedTokens()
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &tc.validMsg)
	s.Require().ErrorContains(err, "epoch tracker not found for epoch stride_epoch")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_HzuNotStatusTransferred() {
	tc := s.SetupClaimUndelegatedTokens()

	// update the hzu status to not transferred
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.validMsg.Epoch)
	s.Require().True(found, "epoch unbonding record found")
	updatedHzu := epochUnbondingRecord.HostZoneUnbondings[0]
	updatedHzu.Status = recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
	newEpochUnbondingRecord, err := s.App.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(s.Ctx, tc.validMsg.Epoch, tc.validMsg.HostZoneId, *updatedHzu)
	s.Require().NoError(err, "epoch unbonding record updated")
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, newEpochUnbondingRecord)

	_, err = s.GetMsgServer().ClaimUndelegatedTokens(s.Ctx, &tc.validMsg)
	s.Require().ErrorContains(err, "host zone unbonding has status: EXIT_TRANSFER_QUEUE, requires status CLAIMABLE")
}
