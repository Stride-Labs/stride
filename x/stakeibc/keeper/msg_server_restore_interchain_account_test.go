package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
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

type LSMTokenDepositStatusUpdate struct {
	chainId        string
	denom          string
	initialStatus  recordtypes.LSMTokenDeposit_Status
	revertedStatus recordtypes.LSMTokenDeposit_Status
}

type RestoreInterchainAccountTestCase struct {
	validMsg                    types.MsgRestoreInterchainAccount
	depositRecordStatusUpdates  []DepositRecordStatusUpdate
	unbondingRecordStatusUpdate []HostZoneUnbondingStatusUpdate
	lsmTokenDepositStatusUpdate []LSMTokenDepositStatusUpdate
	delegationChannelID         string
	delegationPortID            string
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount(createDelegationICAChannel bool) RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	// We have to setup the ICA channel before the LSM Token is stored,
	// otherwise when the EndBlocker runs in the channel setup, the LSM Token
	// statuses will get updated
	var channelID, portID string
	if createDelegationICAChannel {
		owner := "GAIA.DELEGATION"
		channelID, portID = s.CreateICAChannel(owner)
	}

	hostZone := types.HostZone{
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

	// Store LSM Token Deposits with some state pending
	lsmTokenDeposits := []LSMTokenDepositStatusUpdate{
		{
			// Status doesn't change
			chainId:        HostChainId,
			denom:          "denom-1",
			initialStatus:  recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			chainId:        HostChainId,
			denom:          "denom-2",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		},
		{
			// Status doesn't change
			chainId:        HostChainId,
			denom:          "denom-3",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		},
		{
			// Status doesn't change (different host zone)
			chainId:        "different_host_zone",
			denom:          "denom-4",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
		},
	}
	for _, lsmTokenDeposit := range lsmTokenDeposits {
		s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordtypes.LSMTokenDeposit{
			ChainId: lsmTokenDeposit.chainId,
			Status:  lsmTokenDeposit.initialStatus,
			Denom:   lsmTokenDeposit.denom,
		})
	}

	defaultMsg := types.MsgRestoreInterchainAccount{
		Creator:     "creatoraddress",
		ChainId:     HostChainId,
		AccountType: types.ICAAccountType_DELEGATION,
	}

	return RestoreInterchainAccountTestCase{
		validMsg:                    defaultMsg,
		depositRecordStatusUpdates:  depositRecords,
		unbondingRecordStatusUpdate: hostZoneUnbondingRecords,
		lsmTokenDepositStatusUpdate: lsmTokenDeposits,
		delegationChannelID:         channelID,
		delegationPortID:            portID,
	}
}

// Helper function to close an ICA channel
func (s *KeeperTestSuite) closeICAChannel(portId, channelID string) {
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portId, channelID)
	s.Require().True(found, "unable to close channel because channel was not found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, channelID, channel)
}

// Helper function to call RestoreChannel and check that a new channel was created and opened
func (s *KeeperTestSuite) restoreChannelAndVerifySuccess(msg types.MsgRestoreInterchainAccount, portID string, channelID string) {
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

// Helper function to check that each DepositRecord's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyDepositRecordsStatus(expectedDepositRecords []DepositRecordStatusUpdate, revert bool) {
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

// Helper function to check that each HostZoneUnbonding's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyHostZoneUnbondingStatus(expectedUnbondingRecords []HostZoneUnbondingStatusUpdate, revert bool) {
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

// Helper function to check that each LSMTokenDepoit's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyLSMDepositStatus(expectedLSMDeposits []LSMTokenDepositStatusUpdate, revert bool) {
	for i, expectedLSMDeposit := range expectedLSMDeposits {
		actualLSMDeposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, expectedLSMDeposit.chainId, expectedLSMDeposit.denom)
		s.Require().True(found, "lsm deposit found")

		// Only revert record if the revert option is passed and the host zone matches
		expectedStatus := expectedLSMDeposit.initialStatus
		if revert && actualLSMDeposit.ChainId == HostChainId {
			expectedStatus = expectedLSMDeposit.revertedStatus
		}
		s.Require().Equal(expectedStatus.String(), actualLSMDeposit.Status.String(), "lsm deposit %d", i)
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount(true)

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + delegate)")

	// Close the delegation channel
	s.closeICAChannel(tc.delegationPortID, tc.delegationChannelID)

	// Confirm the new channel was created
	s.restoreChannelAndVerifySuccess(tc.validMsg, tc.delegationPortID, tc.delegationChannelID)

	// Verify the record status' were reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, true)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, true)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, true)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_InvalidConnectionId() {
	tc := s.SetupRestoreInterchainAccount(false)

	// Update the connectionId on the host zone so that it doesn't exist
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.ChainId)
	s.Require().True(found)
	hostZone.ConnectionId = "fake_connection"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "invalid connection id from host GAIA, fake_connection not found: invalid request")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount(false)
	msg := tc.validMsg
	msg.AccountType = types.ICAAccountType_WITHDRAWAL

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "ICA controller account address not found: GAIA.WITHDRAWAL")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsForIncorrectHostZone() {
	tc := s.SetupRestoreInterchainAccount(false)
	invalidMsg := tc.validMsg
	invalidMsg.ChainId = "incorrectchainid"

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "host zone not registered")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_RevertDepositRecords_Failure() {
	tc := s.SetupRestoreInterchainAccount(true)

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "existing active channel channel-1 for portID icacontroller-GAIA.DELEGATION")

	// Verify the record status' were NOT reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, false)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_NoRecordChange_Success() {
	// Here, we're closing and restoring the withdrawal channel so records should not be reverted
	tc := s.SetupRestoreInterchainAccount(false)
	owner := "GAIA.WITHDRAWAL"
	channelID, portID := s.CreateICAChannel(owner)

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + withdrawal)")

	// Close the withdrawal channel
	s.closeICAChannel(portID, channelID)

	// Restore the channel
	msg := tc.validMsg
	msg.AccountType = types.ICAAccountType_WITHDRAWAL
	s.restoreChannelAndVerifySuccess(msg, portID, channelID)

	// Verify the record status' were NOT reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, false)
}
