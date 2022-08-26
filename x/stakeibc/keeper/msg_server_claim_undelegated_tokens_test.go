package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type ClaimUndelegatedState struct {
	hostZone           stakeibc.HostZone
	redemptionRecordId string
	redemptionRecord   recordtypes.UserRedemptionRecord
}

type ClaimUndelegatedTestCase struct {
	validMsg       stakeibc.MsgClaimUndelegatedTokens
	initialState   ClaimUndelegatedState
	expectedIcaMsg stakeibckeeper.IcaTx
}

func (s *KeeperTestSuite) SetupClaimUndelegatedTokens() ClaimUndelegatedTestCase {
	redemptionIcaOwner := "GAIA.REDEMPTION"
	s.CreateICAChannel(redemptionIcaOwner)

	epochNumber := uint64(1)
	senderAddr := "stride_SENDER"
	receiverAddr := "cosmos_RECEIVER"
	redemptionAddr := s.IcaAddresses[redemptionIcaOwner]
	redemptionRecordId := fmt.Sprintf("%s.%d.%s", HostChainId, epochNumber, senderAddr)

	redemptionAccount := stakeibc.ICAAccount{
		Address: redemptionAddr,
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	hostZone := stakeibc.HostZone{
		ChainId:           HostChainId,
		RedemptionAccount: &redemptionAccount,
		ConnectionId:      ibctesting.FirstConnectionID,
	}

	redemptionRecord := recordtypes.UserRedemptionRecord{
		Id:          redemptionRecordId,
		HostZoneId:  HostChainId,
		EpochNumber: epochNumber,
		Sender:      senderAddr,
		Receiver:    receiverAddr,
		Denom:       "uatom",
		IsClaimable: true,
		Amount:      1000,
	}
	redemptionAmount := sdk.NewCoins(sdk.NewInt64Coin(redemptionRecord.Denom, int64(redemptionRecord.Amount)))

	epochTracker := stakeibc.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), epochTracker)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), redemptionRecord)

	return ClaimUndelegatedTestCase{
		validMsg: stakeibc.MsgClaimUndelegatedTokens{
			Creator:    senderAddr,
			HostZoneId: HostChainId,
			Epoch:      1,
			Sender:     senderAddr,
		},
		initialState: ClaimUndelegatedState{
			hostZone:           hostZone,
			redemptionRecordId: redemptionRecordId,
			redemptionRecord:   redemptionRecord,
		},
		expectedIcaMsg: stakeibckeeper.IcaTx{
			Msgs: []sdk.Msg{&banktypes.MsgSend{
				FromAddress: redemptionAccount.Address,
				ToAddress:   receiverAddr,
				Amount:      redemptionAmount,
			}},
			Account: redemptionAccount,
			Timeout: uint64(types.DefaultICATimeoutNanos),
		},
	}
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_Successful() {
	tc := s.SetupClaimUndelegatedTokens()
	redemptionRecordId := tc.initialState.redemptionRecordId
	expectedRedemptionRecord := tc.initialState.redemptionRecord

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &tc.validMsg)
	s.Require().NoError(err, "claim undelegated tokens")

	actualRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx(), redemptionRecordId)
	s.Require().True(found, "redemption record found")
	s.Require().False(actualRedemptionRecord.IsClaimable, "redemption record should not be claimable")
	s.Require().Equal(expectedRedemptionRecord.Amount, actualRedemptionRecord.Amount, "redemption record should not be claimable")
	// TODO: check callback data here
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_SuccessfulMsgSendICA() {
	tc := s.SetupClaimUndelegatedTokens()
	redemptionRecord := tc.initialState.redemptionRecord

	icaTx, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx(), &redemptionRecord, redemptionRecord.HostZoneId)
	msgs := icaTx.Msgs
	s.Require().NoError(err, "get redemption transfer msgs error")
	s.Require().Equal(1, len(msgs), "number of transfer messages")
	s.Require().Equal(tc.expectedIcaMsg.Msgs, msgs, "transfer message")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoUserRedemptionRecord() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove the user redemption record
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx(), tc.initialState.redemptionRecordId)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &tc.validMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "could not get user redemption record: GAIA.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_RecordNotClaimable() {
	tc := s.SetupClaimUndelegatedTokens()
	// Mark redemption record as not claimable
	alreadyClaimedRedemptionRecord := tc.initialState.redemptionRecord
	alreadyClaimedRedemptionRecord.IsClaimable = false
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), alreadyClaimedRedemptionRecord)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &tc.validMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "user redemption record is not claimable: GAIA.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_RecordNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	invalidMsg := tc.validMsg
	invalidMsg.HostZoneId = "fake_host_zone"

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &invalidMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "could not get user redemption record: fake_host_zone.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_HostZoneNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	invalidMsg := tc.validMsg
	invalidMsg.HostZoneId = "fake_host_zone"

	badRedemptionRecordId := strings.Replace(tc.initialState.redemptionRecordId, "GAIA", "fake_host_zone", 1)
	badRedemptionRecord := tc.initialState.redemptionRecord
	badRedemptionRecord.Id = badRedemptionRecordId
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), badRedemptionRecord)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &invalidMsg)
	expectedErr := "unable to build redemption transfer message: Host zone fake_host_zone not found: host zone not registered"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoRedemptionAccount() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove redemption account from host zone
	hostZone := tc.initialState.hostZone
	hostZone.RedemptionAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &tc.validMsg)
	expectedErr := "unable to build redemption transfer message: "
	expectedErr += "Redemption account not found for host zone GAIA: host zone not registered"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokens_NoEpochTracker() {
	// Remove epoch tracker
	tc := s.SetupClaimUndelegatedTokens()
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx(), epochtypes.STRIDE_EPOCH)

	_, err := s.GetMsgServer().ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx()), &tc.validMsg)
	expectedErr := "unable to build redemption transfer message: "
	expectedErr += "Epoch tracker not found for epoch stride_epoch: epoch not found"
	s.Require().EqualError(err, expectedErr)
}
