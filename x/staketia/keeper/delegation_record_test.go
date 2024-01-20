package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) addDelegationRecords() (delegationRecords []types.DelegationRecord) {
	for i := 0; i <= 4; i++ {
		delegationRecord := types.DelegationRecord{
			Id:           uint64(i),
			NativeAmount: sdkmath.NewInt(int64(i) * 1000),
		}
		delegationRecords = append(delegationRecords, delegationRecord)
		s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	}
	return delegationRecords
}

func (s *KeeperTestSuite) TestGetDelegationRecord() {
	delegationRecords := s.addDelegationRecords()

	for i := 0; i < len(delegationRecords); i++ {
		expectedRecord := delegationRecords[i]
		recordId := expectedRecord.Id

		actualRecord, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, recordId)
		s.Require().True(found, "delegation record %d should have been found", i)
		s.Require().Equal(expectedRecord, actualRecord)
	}
}

// Tests ArchiveDelegationRecord and GetAllArchivedDelegationRecords
func (s *KeeperTestSuite) TestArchiveDelegationRecord() {
	delegationRecords := s.addDelegationRecords()

	for removedIndex := 0; removedIndex < len(delegationRecords); removedIndex++ {
		// Remove from removed index
		removedId := delegationRecords[removedIndex].Id
		s.App.StaketiaKeeper.ArchiveDelegationRecord(s.Ctx, removedId)

		// Confirm removed
		_, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, removedId)
		s.Require().False(found, "record %d should have been removed", removedId)

		// Check all other records are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(delegationRecords); checkedIndex++ {
			checkedId := delegationRecords[checkedIndex].Id
			_, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, checkedId)
			s.Require().True(found, "record %d should still be here after %d removal", checkedId, removedId)
		}
	}

	// Check that they were all archived
	archivedRecords := s.App.StaketiaKeeper.GetAllArchivedDelegationRecords(s.Ctx)
	for i := 0; i < len(delegationRecords); i++ {
		expectedRecordId := delegationRecords[i].Id
		s.Require().Equal(expectedRecordId, archivedRecords[i].Id, "archived record %d", i)
	}
}

func (s *KeeperTestSuite) TestGetAllActiveDelegationRecords() {
	expectedRecords := s.addDelegationRecords()
	actualRecords := s.App.StaketiaKeeper.GetAllActiveDelegationRecords(s.Ctx)
	s.Require().Equal(len(expectedRecords), len(actualRecords), "number of delegation records")
	s.Require().Equal(expectedRecords, actualRecords)
}

func (s *KeeperTestSuite) TestUpdateDelegationRecordStatus() {
	statuses := []types.DelegationRecordStatus{
		types.TRANSFER_IN_PROGRESS,
		types.DELEGATION_QUEUE,
		types.DELEGATION_ARCHIVE,
	}

	// Create an initial record
	recordId := uint64(1)
	s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, types.DelegationRecord{
		Id: recordId,
	})

	// Iterate through all records and confirm their status updates
	for _, expectedStatus := range statuses {
		err := s.App.StaketiaKeeper.UpdateDelegationRecordStatus(s.Ctx, recordId, expectedStatus)
		s.Require().NoError(err, "no error expected when updating record status")

		delegationRecord, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, recordId)
		s.Require().True(found, "delegation record should have been found")
		s.Require().Equal(expectedStatus, delegationRecord.Status,
			"delegation record status should have been updated to %s", expectedStatus.String())
	}

	// Check that an invalid ID errors
	invalidRecordId := uint64(99)
	err := s.App.StaketiaKeeper.UpdateDelegationRecordStatus(s.Ctx, invalidRecordId, types.DELEGATION_QUEUE)
	s.Require().ErrorContains(err, "delegation record not found")
}
