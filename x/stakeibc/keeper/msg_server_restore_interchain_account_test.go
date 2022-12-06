package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type DepositRecordStatusUpdate struct {
	chainId        string
	initialStatus  recordtypes.DepositRecord_Status
	revertedStatus recordtypes.DepositRecord_Status
}

type HostZoneUnbondingStatusUpdate struct {
	initialStatus  recordtypes.HostZoneUnbonding_Status
	revertedStatus recordtypes.HostZoneUnbonding_Status
}
type RestoreInterchainAccountTestCase struct {
	validMsg                    stakeibc.MsgRestoreInterchainAccount
	depositRecordStatusUpdates  []DepositRecordStatusUpdate
	unbondingRecordStatusUpdate []HostZoneUnbondingStatusUpdate
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount() RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	hostZone := stakeibc.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		RedemptionRate: sdk.OneDec(), // if not yet, the beginblocker invariant panics
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store deposit records with some in state pending
	depositRecords := []DepositRecordStatusUpdate{
		{
			// Status doesn't change
			chainId:        HostChainId,
			initialStatus:  recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			chainId:        HostChainId,
			initialStatus:  recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_DELEGATION_QUEUE,
		},
		{
			// Status doesn't get reveted because it's a different host zone
			chainId:        "different_host_zone",
			initialStatus:  recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
		},
	}
	for i, depositRecord := range depositRecords {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordtypes.DepositRecord{
			Id:         uint64(i),
			HostZoneId: depositRecord.chainId,
			Status:     depositRecord.initialStatus,
		})
	}

	// Store epoch unbonding records with some in state pending
	hostZoneUnbondingRecords := []HostZoneUnbondingStatusUpdate{
		{
			// Status doesn't change
			initialStatus:  recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
			revertedStatus: recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			initialStatus:  recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
			revertedStatus: recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
		{
			// Status doesn't change
			initialStatus:  recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
			revertedStatus: recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			initialStatus:  recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		},
	}
	for i, hostZoneUnbonding := range hostZoneUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: uint64(i),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				// The first unbonding record will get reverted, the other one will not
				{
					HostZoneId: HostChainId,
					Status:     hostZoneUnbonding.initialStatus,
				},
				{
					HostZoneId: "different_host_zone",
					Status:     hostZoneUnbonding.initialStatus,
				},
			},
		})
	}

	defaultMsg := stakeibc.MsgRestoreInterchainAccount{
		Creator:     "creatoraddress",
		ChainId:     HostChainId,
		AccountType: stakeibc.ICAAccountType_DELEGATION,
	}

	return RestoreInterchainAccountTestCase{
		validMsg:                    defaultMsg,
		depositRecordStatusUpdates:  depositRecords,
		unbondingRecordStatusUpdate: hostZoneUnbondingRecords,
	}
}

func (s *KeeperTestSuite) RestoreChannelAndVerifySuccess(msg stakeibc.MsgRestoreInterchainAccount, portID string, channelID string) {
	// Restore the channel
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "registered ica account successfully")

	// Confirm channel was created
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
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

func (s *KeeperTestSuite) VerifyDepositRecordsStatus(expectedDepositRecords []DepositRecordStatusUpdate, revert bool) {
	for i, expectedDepositRecord := range expectedDepositRecords {
		actualDepositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, uint64(i))
		s.Require().True(found, "deposit record found")

		// Only revert records if the revert option is passed and the host zone matches
		expectedStatus := expectedDepositRecord.initialStatus
		if revert && actualDepositRecord.HostZoneId == HostChainId {
			expectedStatus = expectedDepositRecord.revertedStatus
		}
		s.Require().Equal(expectedStatus.String(), actualDepositRecord.Status.String(), "deposit record %d status", i)
	}
}

func (s *KeeperTestSuite) VerifyHostZoneUnbondingStatus(expectedUnbondingRecords []HostZoneUnbondingStatusUpdate, revert bool) {
	for i, expectedUnbonding := range expectedUnbondingRecords {
		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(i))
		s.Require().True(found, "epoch unbonding record found")

		for _, actualUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			// Only revert records if the revert option is passed and the host zone matches
			expectedStatus := expectedUnbonding.initialStatus
			if revert && actualUnbonding.HostZoneId == HostChainId {
				expectedStatus = expectedUnbonding.revertedStatus
			}
			s.Require().Equal(expectedStatus.String(), actualUnbonding.Status.String(), "host zone unbonding for epoch %d record status", i)
		}
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount()
	owner := "GAIA.DELEGATION"
	channelID := s.CreateICAChannel(owner)
	portID := icatypes.PortPrefix + owner

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + delegate)")

	// Close the delegation channel
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portID, channelID)
	s.Require().True(found, "delegation channel found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portID, channelID, channel)

	// Confirm the new channel was created
	s.RestoreChannelAndVerifySuccess(tc.validMsg, portID, channelID)

	// Verify the record status' were reverted
	s.VerifyDepositRecordsStatus(tc.depositRecordStatusUpdates, true)
	s.VerifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, true)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.AccountType = stakeibc.ICAAccountType_WITHDRAWAL

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMSg := fmt.Sprintf("ICA controller account address not found: %s.WITHDRAWAL: invalid interchain account address",
		tc.validMsg.ChainId)
	s.Require().EqualError(err, expectedErrMSg, "registered ica account successfully")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsForIncorrectHostZone() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.ChainId = "incorrectchainid"

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := "host zone not registered"
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails for incorrect host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsIfAccountExists() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.DELEGATION")
	msg := tc.validMsg

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
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

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := fmt.Sprintf("existing active channel channel-1 for portID icacontroller-%s.DELEGATION on connection %s for owner %s.DELEGATION: active channel already set for this owner",
		tc.validMsg.ChainId,
		s.TransferPath.EndpointB.ConnectionID,
		tc.validMsg.ChainId,
	)
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails when account already exists")

	// Verify the record status' were NOT reverted
	s.VerifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.VerifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_NoRecordChange_Success() {
	// Here, we're closing and restoring the withdrawal channel so records should not be reverted
	tc := s.SetupRestoreInterchainAccount()
	owner := "GAIA.WITHDRAWAL"
	channelID := s.CreateICAChannel(owner)
	portID := icatypes.PortPrefix + owner

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + withdrawal)")

	// Close the withdrawal channel
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portID, channelID)
	s.Require().True(found, "withdrawal channel found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portID, channelID, channel)

	// Restore the channel
	msg := tc.validMsg
	msg.AccountType = stakeibc.ICAAccountType_WITHDRAWAL
	s.RestoreChannelAndVerifySuccess(msg, portID, channelID)

	// Verify the record status' were NOT reverted
	s.VerifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.VerifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
}
