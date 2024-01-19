package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) addRedemptionRecords() (redemptionRecords []types.RedemptionRecord) {
	for i := 0; i <= 4; i++ {
		redemptionRecord := types.RedemptionRecord{
			UnbondingRecordId: uint64(i),
			NativeAmount:      sdkmath.NewInt(int64(i) * 1000),
			StTokenAmount:     sdkmath.NewInt(int64(i) * 1000),
			Redeemer:          fmt.Sprintf("address-%d", i),
		}
		redemptionRecords = append(redemptionRecords, redemptionRecord)
		s.App.StakeTiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}
	return redemptionRecords
}

func (s *KeeperTestSuite) TestGetRedemptionRecord() {
	redemptionRecords := s.addRedemptionRecords()

	for i := 0; i < len(redemptionRecords); i++ {
		expectedRecord := redemptionRecords[i]
		unbondingRecordId := expectedRecord.UnbondingRecordId
		redeemer := expectedRecord.Redeemer

		actualRecord, found := s.App.StakeTiaKeeper.GetRedemptionRecord(s.Ctx, unbondingRecordId, redeemer)
		s.Require().True(found, "redemption record %d should have been found", i)
		s.Require().Equal(expectedRecord, actualRecord)
	}
}

func (s *KeeperTestSuite) TestRemoveRedemptionRecord() {
	redemptionRecords := s.addRedemptionRecords()

	for removedIndex := 0; removedIndex < len(redemptionRecords); removedIndex++ {
		// Remove from removed index
		removedRecord := redemptionRecords[removedIndex]
		removedUnbondingId := removedRecord.UnbondingRecordId
		removedRedeemer := removedRecord.Redeemer
		s.App.StakeTiaKeeper.RemoveRedemptionRecord(s.Ctx, removedUnbondingId, removedRedeemer)

		// Confirm removed
		_, found := s.App.StakeTiaKeeper.GetRedemptionRecord(s.Ctx, removedUnbondingId, removedRedeemer)
		s.Require().False(found, "record %d %s should have been removed", removedUnbondingId, removedRedeemer)

		// Check all other records are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(redemptionRecords); checkedIndex++ {
			checkedRecord := redemptionRecords[checkedIndex]
			checkedUnbondingId := checkedRecord.UnbondingRecordId
			checkedRedeemer := checkedRecord.Redeemer

			_, found := s.App.StakeTiaKeeper.GetRedemptionRecord(s.Ctx, checkedUnbondingId, checkedRedeemer)
			s.Require().True(found, "record %d %s should have been removed after %d %s removal",
				checkedUnbondingId, checkedRedeemer, removedUnbondingId, removedRedeemer)
		}
	}
}

func (s *KeeperTestSuite) TestGetAllRedemptionRecord() {
	expectedRecords := s.addRedemptionRecords()
	actualRecords := s.App.StakeTiaKeeper.GetAllRedemptionRecords(s.Ctx)
	s.Require().Equal(len(expectedRecords), len(actualRecords), "number of redemption records")
	s.Require().Equal(expectedRecords, actualRecords)
}

func (s *KeeperTestSuite) TestGetAllRedemptionRecordsFromUnbondingId() {
	// Define a set of redemption records across different unbonding record IDs
	unbondingIdToRecords := map[uint64][]types.RedemptionRecord{
		1: {
			{UnbondingRecordId: 1, Redeemer: "address-A"},
			{UnbondingRecordId: 1, Redeemer: "address-B"},
			{UnbondingRecordId: 1, Redeemer: "address-C"},
		},
		2: {
			{UnbondingRecordId: 2, Redeemer: "address-D"},
			{UnbondingRecordId: 2, Redeemer: "address-E"},
			{UnbondingRecordId: 2, Redeemer: "address-F"},
		},
		3: {
			{UnbondingRecordId: 3, Redeemer: "address-G"},
			{UnbondingRecordId: 3, Redeemer: "address-H"},
			{UnbondingRecordId: 3, Redeemer: "address-I"},
		},
	}

	// Store all the redemption records
	for _, redemptionRecords := range unbondingIdToRecords {
		for _, redemptionRecord := range redemptionRecords {
			s.App.StakeTiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
		}
	}

	// Lookup records by unbonding Id and confirm it matches the expected list
	for unbondingRecordId, expectedRedemptionRecords := range unbondingIdToRecords {
		actualRedemptionRecords := s.App.StakeTiaKeeper.GetRedemptionRecordsFromUnbondingId(s.Ctx, unbondingRecordId)
		s.Require().Equal(len(expectedRedemptionRecords), len(actualRedemptionRecords),
			"number of redemption records for unbonding id %d", unbondingRecordId)

		for i, expectedRecord := range expectedRedemptionRecords {
			actualRecord := actualRedemptionRecords[i]
			s.Require().Equal(expectedRecord.Redeemer, actualRecord.Redeemer, "redemption record address")
			s.Require().Equal(expectedRecord.UnbondingRecordId, actualRecord.UnbondingRecordId,
				"redemption record unbonding ID for %s", expectedRecord.Redeemer)
		}
	}
}
