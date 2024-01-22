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

// Tests that records are written to their respective stores based on the status
func (s *KeeperTestSuite) TestSetDelegationRecord() {
	delegationRecords := []types.DelegationRecord{
		{Id: 1, Status: types.TRANSFER_IN_PROGRESS},
		{Id: 2, Status: types.DELEGATION_QUEUE},
		{Id: 3, Status: types.DELEGATION_ARCHIVE},
		{Id: 4, Status: types.TRANSFER_IN_PROGRESS},
		{Id: 5, Status: types.DELEGATION_QUEUE},
		{Id: 6, Status: types.DELEGATION_ARCHIVE},
	}
	for _, delegationRecord := range delegationRecords {
		s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	}

	// Confirm the number of records in each store
	s.Require().Len(s.App.StaketiaKeeper.GetAllActiveDelegationRecords(s.Ctx), 4, "records in active store")
	s.Require().Len(s.App.StaketiaKeeper.GetAllArchivedDelegationRecords(s.Ctx), 2, "records in archive store")

	// Check that only the non-archived records are found in the active store
	for i, delegationRecord := range delegationRecords {
		expectedFound := delegationRecord.Status != types.DELEGATION_ARCHIVE
		_, actualFound := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, delegationRecord.Id)
		s.Require().Equal(expectedFound, actualFound, "record %d found in active store", i)
	}
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
