package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) addUnbondingRecords() (unbondingRecords []types.UnbondingRecord) {
	for i := 0; i <= 4; i++ {
		unbondingRecord := types.UnbondingRecord{
			Id:            uint64(i),
			NativeAmount:  sdkmath.NewInt(int64(i) * 1000),
			StTokenAmount: sdkmath.NewInt(int64(i) * 1000),
		}
		unbondingRecords = append(unbondingRecords, unbondingRecord)
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}
	return unbondingRecords
}

func (s *KeeperTestSuite) TestGetUnbondingRecord() {
	unbondingRecords := s.addUnbondingRecords()

	for i := 0; i < len(unbondingRecords); i++ {
		expectedRecord := unbondingRecords[i]
		recordId := expectedRecord.Id

		actualRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, recordId)
		s.Require().True(found, "unbonding record %d should have been found", i)
		s.Require().Equal(expectedRecord, actualRecord)
	}
}

// Tests ArchiveUnbondingRecord and GetAllArchivedUnbondingRecords
func (s *KeeperTestSuite) TestArchiveUnbondingRecord() {
	unbondingRecords := s.addUnbondingRecords()

	for removedIndex := 0; removedIndex < len(unbondingRecords); removedIndex++ {
		// Remove from removed index
		removedId := unbondingRecords[removedIndex].Id
		s.App.StaketiaKeeper.ArchiveUnbondingRecord(s.Ctx, removedId)

		// Confirm removed
		_, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, removedId)
		s.Require().False(found, "record %d should have been removed", removedId)

		// Check all other records are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(unbondingRecords); checkedIndex++ {
			checkedId := unbondingRecords[checkedIndex].Id
			_, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, checkedId)
			s.Require().True(found, "record %d should have been removed after %d removal", checkedId, removedId)
		}
	}

	// Check that they were all archived
	archivedRecords := s.App.StaketiaKeeper.GetAllArchivedUnbondingRecords(s.Ctx)
	for i := 0; i < len(unbondingRecords); i++ {
		expectedRecordId := unbondingRecords[i].Id
		s.Require().Equal(expectedRecordId, archivedRecords[i].Id, "archived record %d", i)
	}
}

func (s *KeeperTestSuite) TestGetAllActiveUnbondingRecords() {
	expectedRecords := s.addUnbondingRecords()
	actualRecords := s.App.StaketiaKeeper.GetAllActiveUnbondingRecords(s.Ctx)
	s.Require().Equal(len(expectedRecords), len(actualRecords), "number of unbonding records")
	s.Require().Equal(expectedRecords, actualRecords)
}

func (s *KeeperTestSuite) TestGetAccumulatingUnbondingRecord() {
	expectedRecordId := uint64(3)

	// Set a few records in the store
	unbondingRecords := []types.UnbondingRecord{
		{Id: 1, Status: types.UNBONDING_QUEUE},
		{Id: 2, Status: types.UNBONDING_IN_PROGRESS},
		{Id: 3, Status: types.ACCUMULATING_REDEMPTIONS},
		{Id: 4, Status: types.UNBONDED},
	}
	for _, unbondingRecord := range unbondingRecords {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// Confirm we find the relevant one
	actualAccumulatingRecord, err := s.App.StaketiaKeeper.GetAccumulatingUnbondingRecord(s.Ctx)
	s.Require().NoError(err, "no error expected when grabbing accumulating record")
	s.Require().Equal(expectedRecordId, actualAccumulatingRecord.Id, "found different record than expected")

	// Create an extra ACCUMULATING record and check that it causes an error upon lookup
	duplicateAccumulatingRecordId := uint64(5)
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, types.UnbondingRecord{
		Id:     duplicateAccumulatingRecordId,
		Status: types.ACCUMULATING_REDEMPTIONS,
	})

	_, err = s.App.StaketiaKeeper.GetAccumulatingUnbondingRecord(s.Ctx)
	s.Require().ErrorContains(err, "more than one record")

	// Remove the ACCUMULATING records and confirm it errors
	s.App.StaketiaKeeper.ArchiveUnbondingRecord(s.Ctx, expectedRecordId)
	s.App.StaketiaKeeper.ArchiveUnbondingRecord(s.Ctx, duplicateAccumulatingRecordId)

	_, err = s.App.StaketiaKeeper.GetAccumulatingUnbondingRecord(s.Ctx)
	s.Require().ErrorContains(err, "no unbonding record")
}
