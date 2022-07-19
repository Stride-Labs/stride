package keeper_test

import (
	"fmt"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
}

func (s *KeeperTestSuite) SetupClaimUndelegatedTokens() ClaimUndelegatedTestCase {
	senderAddr := "stride_SENDER"
	receiverAddr := "cosmos_RECEIVER"
	chainId := "GAIA"
	redemptionRecordId := "GAIA.1.stride_SENDER"
	connectionId := "connection-0"

	redemptionAccount := stakeibc.ICAAccount{
		Address: "cosmos_REDEMPTION",
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
	}

	epochTracker := stakeibc.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: s.Ctx.BlockTime().UnixNano() + 100_000_000_000, // plus 100 seconds
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, redemptionRecord)

	connectionEnd := connectiontypes.ConnectionEnd{ClientId: "client-0", Versions: []*connectiontypes.Version{connectiontypes.DefaultIBCVersion}}
	// s.App.ClientState
	// tmClientState, ok := subjectClientState.(*ibctmtypes.ClientState)
	// s.App.GetIBCKeeper().ClientKeeper.GetClientState(s.Ctx, "client-0")
	// store := s.App.IBCKeeper.ClientKeeper.ClientStore(s.Ctx, "client-0")
	// store.Set(host.ClientStateKey(), k.MustMarshalClientState(clientState))
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
	// s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.Wr

	return ClaimUndelegatedTestCase{
		redemptionRecordId: redemptionRecordId,
		validMsg: stakeibc.MsgClaimUndelegatedTokens{
			Creator:    senderAddr,
			HostZoneId: "GAIA",
			Epoch:      1,
			Sender:     senderAddr,
		},
	}
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensSuccessful() {
	tc := s.SetupClaimUndelegatedTokens()
	a, b := s.App.StakeibcKeeper.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, "connection-0", "icacontroller-STRIDE.REDEMPTION")
	fmt.Printf("%v\n", a)
	fmt.Printf("%v\n", b)

	_, err := s.msgServer.ClaimUndelegatedTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err)

	// Confirm the redemption record has been flagged as not claimable
	// redemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, tc.redemptionRecordId)
	// s.Require().True(found)
	// s.Require().False(redemptionRecord.IsClaimable)

	// // Confirm pending claims added
	// pendingClaims := s.App.StakeibcKeeper.GetAllPendingClaims(s.Ctx)
	// s.Require().Equal(1, len(pendingClaims))
	// pendingRedemptionRecordIds := pendingClaims[0].UserRedemptionRecordIds
	// s.Require().Equal(1, len(pendingRedemptionRecordIds))
	// s.Require().Equal(tc.redemptionRecordId, pendingRedemptionRecordIds[0])
	// QUESTION: Anything else to check that I'm missing?
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoUserRedemptionRecord() {
	// tc := s.SetupClaimUndelegatedTokens()
	// // Remove the user redemption record
	// s.App.S

}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensRecordNotClaimable() {
	// Mark redemption record as not claimable
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensHostZoneNotFound() {
	// Change host zone in message
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoRedemptionAccount() {
	// Remove redemption account from host zone
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensNoEpochTracker() {
	// Remove epoch tracker
}

func (s *KeeperTestSuite) TestClaimUndelegatedTokensSubmitTxFailure() {
	// Alter ICA State
}
