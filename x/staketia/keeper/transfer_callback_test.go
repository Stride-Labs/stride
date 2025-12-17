package keeper_test

import (
	"fmt"

	"github.com/Stride-Labs/stride/v31/x/staketia/types"
)

type transferData struct {
	channelId string
	sequence  uint64
	recordId  uint64
}

func (s *KeeperTestSuite) addTransferRecords() (transferRecords []transferData) {
	for i := 0; i <= 4; i++ {
		transferRecord := transferData{
			channelId: fmt.Sprintf("channel-%d", i),
			sequence:  uint64(i),
			recordId:  uint64(i),
		}
		transferRecords = append(transferRecords, transferRecord)
		s.App.StaketiaKeeper.SetTransferInProgressRecordId(s.Ctx, transferRecord.channelId,
			transferRecord.sequence, transferRecord.recordId)
	}
	return transferRecords
}

func (s *KeeperTestSuite) TestGetTransferInProgressRecordId() {
	transferRecords := s.addTransferRecords()

	for i := 0; i < len(transferRecords); i++ {
		expectedRecordId := transferRecords[i].recordId
		channelId := transferRecords[i].channelId
		sequence := transferRecords[i].sequence

		actualRecordId, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, channelId, sequence)
		s.Require().True(found, "redemption record %d should have been found", i)
		s.Require().Equal(expectedRecordId, actualRecordId)
	}
}

func (s *KeeperTestSuite) TestRemoveTransferInProgressRecordId() {
	transferRecords := s.addTransferRecords()

	for removedIndex := 0; removedIndex < len(transferRecords); removedIndex++ {
		// Remove recordId at removed index from store
		removedRecordId := transferRecords[removedIndex].recordId
		removedChannelId := transferRecords[removedIndex].channelId
		removedSequence := transferRecords[removedIndex].sequence
		s.App.StaketiaKeeper.RemoveTransferInProgressRecordId(s.Ctx, removedChannelId, removedSequence)

		// Confirm removed
		_, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, removedChannelId, removedSequence)
		s.Require().False(found, "recordId %d for %s %d should have been removed", removedRecordId, removedChannelId, removedSequence)

		// Check all other recordIds are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(transferRecords); checkedIndex++ {
			checkedRecordId := transferRecords[checkedIndex].recordId
			checkedChannelId := transferRecords[checkedIndex].channelId
			checkedSequence := transferRecords[checkedIndex].sequence

			_, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, checkedChannelId, checkedSequence)
			s.Require().True(found, "recordId %d with %s %d should have been found after %d with %s %d removal",
				checkedRecordId, checkedChannelId, checkedSequence, removedRecordId, removedChannelId, removedSequence)
		}
	}
}

func (s *KeeperTestSuite) TestGetAllTransferInProgressIds() {
	// Store 5 packets across two channels
	expectedTransfers := []types.TransferInProgressRecordIds{}
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			recordId := sequence * 100
			s.App.StaketiaKeeper.SetTransferInProgressRecordId(s.Ctx, channelId, sequence, recordId)
			expectedTransfers = append(expectedTransfers, types.TransferInProgressRecordIds{
				ChannelId: channelId,
				Sequence:  sequence,
				RecordId:  recordId,
			})
		}
	}

	// Check that each transfer is found
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			_, found := s.App.StaketiaKeeper.GetTransferInProgressRecordId(s.Ctx, channelId, sequence)
			s.Require().True(found, "transfer should have been found - channel %s, sequence: %d", channelId, sequence)
		}
	}

	// Check lookup of all transfers
	actualTransfers := s.App.StaketiaKeeper.GetAllTransferInProgressId(s.Ctx)
	s.Require().ElementsMatch(expectedTransfers, actualTransfers, "all transfers")
}
