package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/x/stakedym/types"
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
		s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}
	return redemptionRecords
}

func (s *KeeperTestSuite) TestGetRedemptionRecord() {
	redemptionRecords := s.addRedemptionRecords()

	for i := 0; i < len(redemptionRecords); i++ {
		expectedRecord := redemptionRecords[i]
		unbondingRecordId := expectedRecord.UnbondingRecordId
		redeemer := expectedRecord.Redeemer

		actualRecord, found := s.App.StakedymKeeper.GetRedemptionRecord(s.Ctx, unbondingRecordId, redeemer)
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
		s.App.StakedymKeeper.RemoveRedemptionRecord(s.Ctx, removedUnbondingId, removedRedeemer)

		// Confirm removed
		_, found := s.App.StakedymKeeper.GetRedemptionRecord(s.Ctx, removedUnbondingId, removedRedeemer)
		s.Require().False(found, "record %d %s should have been removed", removedUnbondingId, removedRedeemer)

		// Check all other records are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(redemptionRecords); checkedIndex++ {
			checkedRecord := redemptionRecords[checkedIndex]
			checkedUnbondingId := checkedRecord.UnbondingRecordId
			checkedRedeemer := checkedRecord.Redeemer

			_, found := s.App.StakedymKeeper.GetRedemptionRecord(s.Ctx, checkedUnbondingId, checkedRedeemer)
			s.Require().True(found, "record %d %s should have been removed after %d %s removal",
				checkedUnbondingId, checkedRedeemer, removedUnbondingId, removedRedeemer)
		}
	}
}

func (s *KeeperTestSuite) TestGetAllRedemptionRecord() {
	expectedRecords := s.addRedemptionRecords()
	actualRecords := s.App.StakedymKeeper.GetAllRedemptionRecords(s.Ctx)
	s.Require().Equal(len(expectedRecords), len(actualRecords), "number of redemption records")
	s.Require().ElementsMatch(expectedRecords, actualRecords)
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
			s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
		}
	}

	// Lookup records by unbonding Id and confirm it matches the expected list
	for unbondingRecordId, expectedRedemptionRecords := range unbondingIdToRecords {
		actualRedemptionRecords := s.App.StakedymKeeper.GetRedemptionRecordsFromUnbondingId(s.Ctx, unbondingRecordId)
		s.Require().Equal(len(expectedRedemptionRecords), len(actualRedemptionRecords),
			"number of redemption records for unbonding id %d", unbondingRecordId)

		for i, expectedRecord := range expectedRedemptionRecords {
			actualRecord := actualRedemptionRecords[i]
			s.Require().Equal(expectedRecord.Redeemer, actualRecord.Redeemer, "redemption record address")
			s.Require().Equal(unbondingRecordId, actualRecord.UnbondingRecordId,
				"redemption record unbonding ID for %s", expectedRecord.Redeemer)
		}
	}
}

func (s *KeeperTestSuite) TestGetRedemptionRecordsFromAddress() {
	// Define a set of redemption records across different addresses
	unbondingAddressToRecords := map[string][]types.RedemptionRecord{
		"address-A": {
			{UnbondingRecordId: 1, Redeemer: "address-A"},
			{UnbondingRecordId: 2, Redeemer: "address-A"},
			{UnbondingRecordId: 3, Redeemer: "address-A"},
		},
		"address-B": {
			{UnbondingRecordId: 4, Redeemer: "address-B"},
			{UnbondingRecordId: 5, Redeemer: "address-B"},
			{UnbondingRecordId: 6, Redeemer: "address-B"},
		},
		"address-C": {
			{UnbondingRecordId: 7, Redeemer: "address-C"},
			{UnbondingRecordId: 8, Redeemer: "address-C"},
			{UnbondingRecordId: 9, Redeemer: "address-C"},
		},
	}

	// Store all the redemption records
	for _, redemptionRecords := range unbondingAddressToRecords {
		for _, redemptionRecord := range redemptionRecords {
			s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
		}
	}

	// Lookup records by address and confirm it matches the expected list
	for expectedAddress, expectedRedemptionRecords := range unbondingAddressToRecords {
		actualRedemptionRecords := s.App.StakedymKeeper.GetRedemptionRecordsFromAddress(s.Ctx, expectedAddress)
		s.Require().Equal(len(expectedRedemptionRecords), len(actualRedemptionRecords),
			"number of redemption records for address %d", expectedAddress)

		for i, expectedRecord := range expectedRedemptionRecords {
			actualRecord := actualRedemptionRecords[i]
			s.Require().Equal(expectedAddress, actualRecord.Redeemer, "redemption record address")
			s.Require().Equal(expectedRecord.UnbondingRecordId, actualRecord.UnbondingRecordId,
				"redemption record unbonding ID for %s", expectedRecord.Redeemer)
		}
	}
}
