package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v26/x/stakedym/types"
)

type PacketCallbackTestCase struct {
	ChannelId        string
	OriginalSequence uint64
	RetrySequence    uint64
	Token            sdk.Coin
	Packet           channeltypes.Packet
	Record           types.DelegationRecord
}

func (s *KeeperTestSuite) SetupTestHandleRecordUpdatePacket() PacketCallbackTestCase {
	senderAccount := s.TestAccs[0]

	// IBC transfer packet data
	sequence := uint64(1)
	channelId := "channel-0"

	// Pending delegation record associated with transfer
	record := types.DelegationRecord{
		Id:           1,
		NativeAmount: sdk.NewInt(0),
		Status:       types.TRANSFER_IN_PROGRESS,
	}

	// Write record to store
	s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, record)

	// Add the pending record to the store
	s.App.StakedymKeeper.SetTransferInProgressRecordId(s.Ctx, channelId, sequence, record.Id)

	// Build the IBC packet
	transferMetadata := transfertypes.FungibleTokenPacketData{
		Denom:  "denom",
		Sender: senderAccount.String(),
	}
	packet := channeltypes.Packet{
		Sequence:      sequence,
		SourceChannel: channelId,
		Data:          transfertypes.ModuleCdc.MustMarshalJSON(&transferMetadata),
	}

	return PacketCallbackTestCase{
		ChannelId:        channelId,
		OriginalSequence: sequence,
		Packet:           packet,
		Record:           record,
	}
}

func (s *KeeperTestSuite) TestArchiveFailedTransferRecord() {
	// Create an initial record
	recordId := uint64(1)
	s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, types.DelegationRecord{
		Id: recordId,
	})

	// Update the hash to failed
	err := s.App.StakedymKeeper.ArchiveFailedTransferRecord(s.Ctx, recordId)
	s.Require().NoError(err, "no error expected when archiving transfer record")

	// Confirm it was updated
	delegationRecord, found := s.App.StakedymKeeper.GetArchivedDelegationRecord(s.Ctx, recordId)
	s.Require().True(found, "delegation record should have been archived")
	s.Require().Equal(types.TRANSFER_FAILED, delegationRecord.Status, "delegation record status")

	// Check that an invalid ID errors
	invalidRecordId := uint64(99)
	err = s.App.StakedymKeeper.ArchiveFailedTransferRecord(s.Ctx, invalidRecordId)
	s.Require().ErrorContains(err, "delegation record not found")
}

// --------------------------------------------------------------
//                        OnTimeoutPacket
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestOnTimeoutPacket_Successful() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Call OnTimeoutPacket
	err := s.App.StakedymKeeper.OnTimeoutPacket(s.Ctx, tc.Packet)
	s.Require().NoError(err, "no error expected when calling OnTimeoutPacket")

	s.verifyDelegationRecordArchived(tc)
}

func (s *KeeperTestSuite) TestOnTimeoutPacket_NoOp() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Get all delegation records
	recordsBefore := s.getAllRecords(tc)

	// Remove the callback data
	s.App.StakedymKeeper.RemoveTransferInProgressRecordId(s.Ctx, tc.ChannelId, tc.OriginalSequence)

	// Should be a no-op since there's no callback data
	err := s.App.StakedymKeeper.OnTimeoutPacket(s.Ctx, tc.Packet)
	s.Require().NoError(err, "no error expected when calling OnTimeoutPacket")

	s.verifyNoRecordsChanged(tc, recordsBefore)
}

// --------------------------------------------------------------
//                    OnAcknowledgementPacket
// --------------------------------------------------------------

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_AckSuccess() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Build a successful ack
	ackSuccess := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{1}, // just has to be non-empty
		},
	})

	// Call OnAckPacket with the successful ack
	err := s.App.StakedymKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, ackSuccess)
	s.Require().NoError(err, "no error expected during OnAckPacket")

	s.verifyDelegationRecordQueued(tc)
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_AckFailure() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Build an error ack
	ackFailure := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{},
	})

	// Call OnAckPacket with the successful ack
	err := s.App.StakedymKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, ackFailure)
	s.Require().NoError(err, "no error expected during OnAckPacket")

	s.verifyDelegationRecordArchived(tc)
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket_InvalidAck() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Get all delegation records
	recordsBefore := s.getAllRecords(tc)

	// Build an invalid ack to force an error
	invalidAck := transfertypes.ModuleCdc.MustMarshalJSON(&channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Result{
			Result: []byte{}, // empty result causes an error
		},
	})

	// Call OnAckPacket with the invalid ack
	err := s.App.StakedymKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, invalidAck)
	s.Require().ErrorContains(err, "invalid acknowledgement")

	// Verify store is unchanged
	s.verifyNoRecordsChanged(tc, recordsBefore)
}

// record not found for record id case
func (s *KeeperTestSuite) TestOnAcknowledgementPacket_NoOp() {
	tc := s.SetupTestHandleRecordUpdatePacket()

	// Get all delegation records
	recordsBefore := s.getAllRecords(tc)

	// Remove the record id so that there is no action necessary in the callback
	s.App.StakedymKeeper.RemoveTransferInProgressRecordId(s.Ctx, tc.ChannelId, tc.OriginalSequence)

	// Call OnAckPacket and confirm there was no error
	// The ack argument here doesn't matter cause the no-op check is upstream
	err := s.App.StakedymKeeper.OnAcknowledgementPacket(s.Ctx, tc.Packet, []byte{})
	s.Require().NoError(err, "no error expected during on ack packet")

	// Verify store is unchanged
	s.verifyNoRecordsChanged(tc, recordsBefore)
}

// --------------------------------------------------------------
//                    Helpers
// --------------------------------------------------------------

// Helper function to verify the record was updated after a successful transfer
func (s *KeeperTestSuite) verifyDelegationRecordQueued(tc PacketCallbackTestCase) {
	// Confirm the DelegationRecord is still in the active store
	record, found := s.App.StakedymKeeper.GetDelegationRecord(s.Ctx, tc.Record.Id)
	s.Require().True(found, "record should have been found")
	// Confirm the record was not archived
	_, found = s.App.StakedymKeeper.GetArchivedDelegationRecord(s.Ctx, tc.Record.Id)
	s.Require().False(found, "record should not be archived")

	// Confirm the record is unchanged, except for the status
	tc.Record.Status = types.DELEGATION_QUEUE
	s.Require().Equal(tc.Record, record, "record should have been archived")

	// Confirm the transfer in progress was removed
	_, found = s.App.StakedymKeeper.GetTransferInProgressRecordId(s.Ctx, tc.ChannelId, tc.OriginalSequence)
	s.Require().False(found, "transfer in progress should have been removed")
}

// Helper function to verify record was archived after a failed or timed out transfer
func (s *KeeperTestSuite) verifyDelegationRecordArchived(tc PacketCallbackTestCase) {
	// Confirm the DelegationRecord was archived
	archivedRecord, found := s.App.StakedymKeeper.GetArchivedDelegationRecord(s.Ctx, tc.Record.Id)
	s.Require().True(found, "record should have been found in the archive store")
	// Confirm the record is no longer in the active store
	_, found = s.App.StakedymKeeper.GetDelegationRecord(s.Ctx, tc.Record.Id)
	s.Require().False(found, "record should have been removed from the store")

	// Confirm the record is unchanged, except for the Status
	tc.Record.Status = types.TRANSFER_FAILED
	s.Require().Equal(tc.Record, archivedRecord, "record should have been archived")

	// Confirm the transfer in progress was removed
	_, found = s.App.StakedymKeeper.GetTransferInProgressRecordId(s.Ctx, tc.ChannelId, tc.OriginalSequence)
	s.Require().False(found, "transfer in progress should have been removed")
}

// Helper function to grab both active and archived delegation records
func (s *KeeperTestSuite) getAllRecords(tc PacketCallbackTestCase) (allRecords []types.DelegationRecord) {
	// Get all delegation records
	activeRecords := s.App.StakedymKeeper.GetAllActiveDelegationRecords(s.Ctx)
	archiveRecords := s.App.StakedymKeeper.GetAllArchivedDelegationRecords(s.Ctx)
	// append the records
	allRecords = append(activeRecords, archiveRecords...)
	return allRecords
}

// Helper function to verify no records were updated
func (s *KeeperTestSuite) verifyNoRecordsChanged(tc PacketCallbackTestCase, recordsBefore []types.DelegationRecord) {
	// Get current records
	recordsAfter := s.getAllRecords(tc)
	// Compare to records before
	s.Require().Equal(recordsBefore, recordsAfter, "records should be unchanged")
}
