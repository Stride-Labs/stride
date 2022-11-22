package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	recordtypes "github.com/Stride-Labs/stride/v3/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

const (
	numDepositRecords             = 3
	numEpochUnbondingRecords      = 3
	startingDepositRecordStatus   = recordtypes.DepositRecord_DELEGATION_IN_PROGRESS
	startingUnbondingRecordStatus = recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS
)

type RestoreInterchainAccountTestCase struct {
	validMsg stakeibc.MsgRestoreInterchainAccount
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount() RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	hostZone := stakeibc.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		RedemptionRate: sdk.OneDec(), // if not yet, the beginblocker invariant panics
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	// Store pending deposit records
	for i := 0; i < numDepositRecords; i++ {
		// Store a different host zone for the first record
		chainId := HostChainId
		if i == 0 {
			chainId = "differentHostZone"
		}
		depositRecord := recordtypes.DepositRecord{
			Id:                 uint64(i),
			DepositEpochNumber: uint64(i),
			HostZoneId:         chainId,
			Status:             startingDepositRecordStatus,
		}
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx(), depositRecord)
	}

	// Store pending epoch unbonding records
	for i := 0; i < numEpochUnbondingRecords; i++ {
		// Store a different host zone for the first record
		epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
			EpochNumber: uint64(i),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId: HostChainId,
					Status:     startingUnbondingRecordStatus,
				},
				{
					HostZoneId: "differentHostZone",
					Status:     startingUnbondingRecordStatus,
				},
			},
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	}

	defaultMsg := stakeibc.MsgRestoreInterchainAccount{
		Creator:     "creatoraddress",
		ChainId:     HostChainId,
		AccountType: stakeibc.ICAAccountType_DELEGATION,
	}

	return RestoreInterchainAccountTestCase{
		validMsg: defaultMsg,
	}
}

func (s *KeeperTestSuite) RestoreChannelAndVerifySuccess(msg stakeibc.MsgRestoreInterchainAccount, portID string, channelID string) {
	// Restore the channel
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "registered ica account successfully")

	// Confirm channel was created
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	s.Require().Len(channels, 3, "there should be 3 channels after restoring")

	// Confirm the new channel is in state INIT
	newChannelActive := false
	for _, channel := range channels {
		// The new channel should have the same port, a new channel ID and be in state INIT
		if channel.PortId == portID && channel.ChannelId != channelID && channel.State == channeltypes.INIT {
			newChannelActive = true
		}
	}
	s.Require().True(newChannelActive, "a new channel should have been created")
}

func (s *KeeperTestSuite) VerifyDepositRecordsStatus(status recordtypes.DepositRecord_Status) {
	for i := 0; i < numDepositRecords; i++ {
		depositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), uint64(i))
		s.Require().True(found, "deposit record found")

		if depositRecord.HostZoneId == HostChainId {
			s.Require().Equal(status.String(), depositRecord.Status.String(), "deposit record %d status", i)
		} else {
			// Any other host chain should not be updated
			s.Require().Equal(startingDepositRecordStatus.String(), depositRecord.Status.String(), "deposit record %d status", i)
		}
	}
}

func (s *KeeperTestSuite) VerifyHostZoneUnbondingStatus(status recordtypes.HostZoneUnbonding_Status) {
	for i := 0; i < numEpochUnbondingRecords; i++ {
		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), uint64(i))
		s.Require().True(found, "epoch unbonding record found")
		for _, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			if hostZoneUnbonding.HostZoneId == HostChainId {
				s.Require().Equal(status.String(), hostZoneUnbonding.Status.String(), "host zone unbonding record status")
			} else {
				// Any other host chain should not be updated
				s.Require().Equal(startingUnbondingRecordStatus.String(), hostZoneUnbonding.Status.String(), "host zone unbonding record status")
			}
		}
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount()
	owner := "GAIA.DELEGATION"
	channelID := s.CreateICAChannel(owner)
	portID := icatypes.PortPrefix + owner

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + delegate)")

	// Close the delegation channel
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx(), portID, channelID)
	s.Require().True(found, "delegation channel found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx(), portID, channelID, channel)

	// Confirm the new channel was created
	s.RestoreChannelAndVerifySuccess(tc.validMsg, portID, channelID)

	// Verify the deposit record state was reverted
	s.VerifyDepositRecordsStatus(recordtypes.DepositRecord_DELEGATION_QUEUE)
	s.VerifyHostZoneUnbondingStatus(recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.AccountType = stakeibc.ICAAccountType_WITHDRAWAL
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMSg := fmt.Sprintf("ICA controller account address not found: %s.WITHDRAWAL: invalid interchain account address",
		tc.validMsg.ChainId)
	s.Require().EqualError(err, expectedErrMSg, "registered ica account successfully")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsForIncorrectHostZone() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.ChainId = "incorrectchainid"
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := "host zone not registered"
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails for incorrect host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsIfAccountExists() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.DELEGATION")
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := fmt.Sprintf("existing active channel channel-1 for portID icacontroller-%s.DELEGATION on connection %s for owner %s.DELEGATION: active channel already set for this owner",
		tc.validMsg.ChainId,
		s.TransferPath.EndpointB.ConnectionID,
		tc.validMsg.ChainId,
	)
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails when account already exists")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_RevertDepositRecords_Failure() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.DELEGATION")
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := fmt.Sprintf("existing active channel channel-1 for portID icacontroller-%s.DELEGATION on connection %s for owner %s.DELEGATION: active channel already set for this owner",
		tc.validMsg.ChainId,
		s.TransferPath.EndpointB.ConnectionID,
		tc.validMsg.ChainId,
	)
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails when account already exists")
	// Verify the deposit record state was NOT reverted
	s.VerifyDepositRecordsStatus(startingDepositRecordStatus)
	s.VerifyHostZoneUnbondingStatus(startingUnbondingRecordStatus)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_NoRecordChange_Success() {
	// Here, we're closing and restoring the withdrawal channel so records should not be reverted
	tc := s.SetupRestoreInterchainAccount()
	owner := "GAIA.WITHDRAWAL"
	channelID := s.CreateICAChannel(owner)
	portID := icatypes.PortPrefix + owner

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + withdrawal)")

	// Close the withdrawal channel
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx(), portID, channelID)
	s.Require().True(found, "withdrawal channel found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx(), portID, channelID, channel)

	// Restore the channel
	msg := tc.validMsg
	msg.AccountType = stakeibc.ICAAccountType_WITHDRAWAL
	s.RestoreChannelAndVerifySuccess(msg, portID, channelID)

	// Verify the records were NOT CHANGED
	s.VerifyDepositRecordsStatus(startingDepositRecordStatus)
	s.VerifyHostZoneUnbondingStatus(startingUnbondingRecordStatus)
}
