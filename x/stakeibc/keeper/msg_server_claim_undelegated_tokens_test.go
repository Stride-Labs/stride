package keeper_test

import (
	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

	// host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	_ "github.com/stretchr/testify/suite"
)

type ClaimUndelegatedTestCase struct {
	redemptionRecordId string
	validMsg           stakeibc.MsgClaimUndelegatedTokens
	expectedTxMsg      stakeibckeeper.IcaTx
	redemptionRecord   recordtypes.UserRedemptionRecord
	hostZone           types.HostZone
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

	connectionEnd := connectiontypes.ConnectionEnd{ClientId: "client-0", Versions: []*connectiontypes.Version{connectiontypes.DefaultIBCVersion}}
	owner := stakeibc.FormatICAAccountOwner("STRIDE", stakeibc.ICAAccountType_REDEMPTION)
	portId := "icacontroller-" + owner
	state := ibctmtypes.NewClientState("STRIDE", ibctmtypes.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clienttypes.NewHeight(0, 10), commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
	s.App.StakeibcKeeper.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, "client-0", state)
	s.App.StakeibcKeeper.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionId, connectionEnd)
	s.App.StakeibcKeeper.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionId, connectionEnd)

	s.App.ICAControllerKeeper.RegisterInterchainAccount(s.Ctx, connectionId, owner)
	s.App.ICAControllerKeeper.SetActiveChannelID(s.Ctx, connectionId, portId, "channel-0")
	channel, _ := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, "connection-0", "icacontroller-STRIDE.REDEMPTION")
	channel.State = channeltypes.OPEN
	channel.Counterparty = channeltypes.NewCounterparty(portId, "channel-0")
	s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, "channel-0", channel)

	msgSend := banktypes.MsgSend{
		FromAddress: redemptionAccount.Address,
		ToAddress:   receiverAddr,
		Amount:      redemptionAmount,
	}

	return ClaimUndelegatedTestCase{
		redemptionRecordId: redemptionRecordId,
		validMsg: stakeibc.MsgClaimUndelegatedTokens{
			Creator:    senderAddr,
			HostZoneId: "GAIA",
			Epoch:      1,
			Sender:     senderAddr,
		},
		expectedTxMsg: stakeibckeeper.IcaTx{
			ConnectionId: connectionId,
			Msgs:         []sdk.Msg{&msgSend},
			Account:      redemptionAccount,
			Timeout:      uint64(types.DefaultICATimeoutNanos),
		},
		redemptionRecord: redemptionRecord,
		hostZone:         hostZone,
	}
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensSuccessful() {
	tc := s.SetupClaimUndelegatedTokens()

	userRedemptionRecord, err := s.App.StakeibcKeeper.GetClaimableRedemptionRecord(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "get redemptions record should not error")

	actualTxMsg, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, userRedemptionRecord, tc.validMsg.HostZoneId)
	s.Require().NoError(err, "get redemption transfer msg should not error")
	s.Require().Equal(tc.expectedTxMsg, *actualTxMsg, "redemption transfer message")

	//Confirm the redemption record has been flagged as not claimable
	s.App.StakeibcKeeper.FlagRedemptionRecordsAsClaimed(s.Ctx, userRedemptionRecord, 1)
	redemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, tc.redemptionRecordId)
	s.Require().True(found)
	s.Require().False(redemptionRecord.IsClaimable)

	// Confirm pending claims added
	pendingClaims := s.App.StakeibcKeeper.GetAllPendingClaims(s.Ctx)
	s.Require().Equal(1, len(pendingClaims))
	pendingRedemptionRecordIds := pendingClaims[0].UserRedemptionRecordIds
	s.Require().Equal(1, len(pendingRedemptionRecordIds))
	s.Require().Equal(tc.redemptionRecordId, pendingRedemptionRecordIds[0])
	// QUESTION: Anything else to check that I'm missing?
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoUserRedemptionRecord() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove the user redemption record
	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, tc.redemptionRecordId)

	_, err := s.App.StakeibcKeeper.GetClaimableRedemptionRecord(s.Ctx, &tc.validMsg)
	s.Require().EqualError(err, "could not get user redemption record: GAIA.1.stride_SENDER: user redemption record error")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensRecordNotClaimable() {
	tc := s.SetupClaimUndelegatedTokens()
	// Mark redemption record as not claimable
	alreadyClaimedRedemptionRecord := tc.redemptionRecord
	alreadyClaimedRedemptionRecord.IsClaimable = false
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, alreadyClaimedRedemptionRecord)

	_, err := s.App.StakeibcKeeper.GetClaimableRedemptionRecord(s.Ctx, &tc.validMsg)
	s.Require().EqualError(err, "user redemption record is not claimable: GAIA.1.stride_SENDER: user redemption record error")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensHostZoneNotFound() {
	tc := s.SetupClaimUndelegatedTokens()
	// Change host zone in message
	_, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &tc.redemptionRecord, "fake_host_zone")
	s.Require().EqualError(err, "Host zone fake_host_zone not found: host zone not registered")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoRedemptionAccount() {
	tc := s.SetupClaimUndelegatedTokens()
	// Remove redemption account from host zone
	hostZone := tc.hostZone
	hostZone.RedemptionAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &tc.redemptionRecord, tc.validMsg.HostZoneId)
	s.Require().EqualError(err, "Redemption account not found for host zone GAIA: host zone not registered")
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoEpochTracker() {
	// Remove epoch tracker
	tc := s.SetupClaimUndelegatedTokens()
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.App.StakeibcKeeper.GetRedemptionTransferMsg(s.Ctx, &tc.redemptionRecord, tc.validMsg.HostZoneId)
	s.Require().EqualError(err, "Epoch tracker not found for epoch stride_epoch: epoch not found")
}
