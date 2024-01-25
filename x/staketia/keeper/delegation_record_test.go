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
			TxHash:       "hash",
		}
		delegationRecords = append(delegationRecords, delegationRecord)
		s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
	}
	return delegationRecords
}

func (s *KeeperTestSuite) TestSafelySetDelegationRecord() {
	// Set one record
	err := s.App.StaketiaKeeper.SafelySetDelegationRecord(s.Ctx, types.DelegationRecord{Id: 1})
	s.Require().NoError(err, "no error expected when setting record")

	// Attempt to set it again, it should fail
	err = s.App.StaketiaKeeper.SafelySetDelegationRecord(s.Ctx, types.DelegationRecord{Id: 1})
	s.Require().ErrorContains(err, "delegation record already exists")

	// Set a new ID, it should succeed
	err = s.App.StaketiaKeeper.SafelySetDelegationRecord(s.Ctx, types.DelegationRecord{Id: 2})
	s.Require().NoError(err, "no error expected when setting new ID")
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
		// Archive from removed index
		removedRecord := delegationRecords[removedIndex]
		s.App.StaketiaKeeper.ArchiveDelegationRecord(s.Ctx, removedRecord)

		// Confirm removed from active
		_, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, removedRecord.Id)
		s.Require().False(found, "record %d should have been removed", removedRecord.Id)

		// Confirm placed in archive
		_, found = s.App.StaketiaKeeper.GetArchivedDelegationRecord(s.Ctx, removedRecord.Id)
		s.Require().True(found, "record %d should have been moved to the archive store", removedRecord.Id)

		// Check all other records are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(delegationRecords); checkedIndex++ {
			checkedId := delegationRecords[checkedIndex].Id
			_, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, checkedId)
			s.Require().True(found, "record %d should still be here after %d removal", checkedId, removedRecord.Id)
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
