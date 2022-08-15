package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	senderAddr := "stride_SENDER"
	receiverAddr := "cosmos_RECEIVER"
	redemptionAddr := "cosmos_REDEMPTION"
	chainId := "GAIA"
	redemptionRecordId := "GAIA.1.stride_SENDER"
	connectionId := "connection-0"

	redemptionAccount := stakeibc.ICAAccount{
		Address: redemptionAddr,
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	hostZone := stakeibc.HostZone{
		ChainId:           chainId,
		RedemptionAccount: &redemptionAccount,
		ConnectionId:      connectionId,
	}

	redemptionRecord := recordtypes.UserRedemptionRecord{
		Id:          redemptionRecordId,
		HostZoneId:  chainId,
		EpochNumber: 1,
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
		NextEpochStartTime: 0,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, redemptionRecord)

	return ClaimUndelegatedTestCase{
		validMsg: stakeibc.MsgClaimUndelegatedTokens{
			Creator:    senderAddr,
			HostZoneId: "GAIA",
			Epoch:      1,
			Sender:     senderAddr,
		},
		initialState: ClaimUndelegatedState{
			hostZone:           hostZone,
			redemptionRecordId: redemptionRecordId,
			redemptionRecord:   redemptionRecord,
		},
		expectedIcaMsg: stakeibckeeper.IcaTx{
			ConnectionId: connectionId,
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

func (s *KeeperTestSuite) TestClaimUndelegatedTokensSuccessful() {
	s.T().Skip("fixing on redemption branch")
	tc := s.SetupClaimUndelegatedTokens()
	redemptionRecordId := tc.initialState.redemptionRecordId

	userRedemptionRecord, err := s.App.StakeibcKeeper.GetClaimableRedemptionRecord(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "get redemptions record should not error")

	actualTxMsg, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, userRedemptionRecord, tc.validMsg.HostZoneId)
	s.Require().NoError(err, "get redemption transfer msg should not error")
	s.Require().Equal(tc.expectedIcaMsg, *actualTxMsg, "redemption transfer message")

	redemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, redemptionRecordId)
	s.Require().True(found)
	s.Require().False(redemptionRecord.IsClaimable)

	// TODO: check callback data here
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoUserRedemptionRecord() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove the user redemption record
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, tc.initialState.redemptionRecordId)

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "could not get user redemption record: GAIA.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensRecordNotClaimable() {
	tc := s.SetupClaimUndelegatedTokens()
	// Mark redemption record as not claimable
	alreadyClaimedRedemptionRecord := tc.initialState.redemptionRecord
	alreadyClaimedRedemptionRecord.IsClaimable = false
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, alreadyClaimedRedemptionRecord)

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "user redemption record is not claimable: GAIA.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensHostZoneNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	invalidMsg := tc.validMsg
	invalidMsg.HostZoneId = "fake_host_zone"

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	expectedErr := "unable to find claimable redemption record: "
	expectedErr += "could not get user redemption record: fake_host_zone.1.stride_SENDER: user redemption record error"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoRedemptionAccount() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove redemption account from host zone
	hostZone := tc.initialState.hostZone
	hostZone.RedemptionAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErr := "unable to build redemption transfer message: "
	expectedErr += "Redemption account not found for host zone GAIA: host zone not registered"
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoEpochTracker() {
	// Remove epoch tracker
	tc := s.SetupClaimUndelegatedTokens()
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErr := "unable to build redemption transfer message: "
	expectedErr += "Epoch tracker not found for epoch stride_epoch: epoch not found"
	s.Require().EqualError(err, expectedErr)
}
