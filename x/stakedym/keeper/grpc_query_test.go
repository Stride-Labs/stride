package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Stride-Labs/stride/v28/x/stakedym/types"
)

func (s *KeeperTestSuite) TestQueryHostZone() {
	chainId := "chain-0"
	hostZone := types.HostZone{
		ChainId: chainId,
	}
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	req := &types.QueryHostZoneRequest{}
	resp, err := s.App.StakedymKeeper.HostZone(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying host zone")
	s.Require().Equal(chainId, resp.HostZone.ChainId, "host zone chain-id from query")
}

func (s *KeeperTestSuite) TestQueryDelegationRecords() {
	// Create active delegation records
	initialDelegationRecords := s.addDelegationRecords()

	// Create an archived version of each of the above records by archiving
	// the record and then recreating it in the new store
	archivedDelegationRecords := []types.DelegationRecord{}
	activeDelegationRecords := []types.DelegationRecord{}
	for _, delegationRecord := range initialDelegationRecords {
		// Update the status and archive teh record
		// (which removes from the active store, and writes to the archive store)
		archivedRecord := delegationRecord
		archivedRecord.Status = types.DELEGATION_COMPLETE
		s.App.StakedymKeeper.ArchiveDelegationRecord(s.Ctx, archivedRecord)
		archivedDelegationRecords = append(archivedDelegationRecords, archivedRecord)

		// Set the original record back to the active store
		delegationRecord.Status = types.TRANSFER_IN_PROGRESS
		s.App.StakedymKeeper.SetDelegationRecord(s.Ctx, delegationRecord)
		activeDelegationRecords = append(activeDelegationRecords, delegationRecord)
	}
	allDelegationRecords := append(activeDelegationRecords, archivedDelegationRecords...)

	// Test a query with no archived records
	activeReq := &types.QueryDelegationRecordsRequest{IncludeArchived: false}
	activeResp, err := s.App.StakedymKeeper.DelegationRecords(s.Ctx, activeReq)
	s.Require().NoError(err, "no error expected when querying active records")

	s.Require().Equal(len(activeDelegationRecords), len(activeResp.DelegationRecords), "number of active records")
	s.Require().ElementsMatch(activeDelegationRecords, activeResp.DelegationRecords, "active records")

	// Test a query with all records (including archived records)
	allReq := &types.QueryDelegationRecordsRequest{IncludeArchived: true}
	allResp, err := s.App.StakedymKeeper.DelegationRecords(s.Ctx, allReq)
	s.Require().NoError(err, "no error expected when querying all records")

	s.Require().Equal(len(allDelegationRecords), len(allResp.DelegationRecords), "all records")
	s.Require().ElementsMatch(allDelegationRecords, allResp.DelegationRecords, "all records")
}

func (s *KeeperTestSuite) TestQueryUnbondingRecords() {
	// Create active unbondin records
	initialUnbondingRecords := s.addUnbondingRecords()

	// Create an archived version of each of the above records by archiving the record
	// and then recreating it in the new store
	archivedUnbondingRecords := []types.UnbondingRecord{}
	activeUnbondingRecords := []types.UnbondingRecord{}
	for _, unbondingRecord := range initialUnbondingRecords {
		// Archive (which removes from the active store, and writes to the archive store)
		archivedRecord := unbondingRecord
		archivedRecord.Status = types.CLAIMED
		s.App.StakedymKeeper.ArchiveUnbondingRecord(s.Ctx, archivedRecord)
		archivedUnbondingRecords = append(archivedUnbondingRecords, archivedRecord)

		// Set the original record back to the active store
		unbondingRecord.Status = types.UNBONDING_QUEUE
		s.App.StakedymKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
		activeUnbondingRecords = append(activeUnbondingRecords, unbondingRecord)
	}
	allUnbondingRecords := append(activeUnbondingRecords, archivedUnbondingRecords...)

	// Test a query with no archived records
	activeReq := &types.QueryUnbondingRecordsRequest{IncludeArchived: false}
	activeResp, err := s.App.StakedymKeeper.UnbondingRecords(s.Ctx, activeReq)
	s.Require().NoError(err, "no error expected when querying active records")

	s.Require().Equal(len(activeUnbondingRecords), len(activeResp.UnbondingRecords), "number of active records")
	s.Require().ElementsMatch(activeUnbondingRecords, activeResp.UnbondingRecords, "active records")

	// Test a query with no all records
	allReq := &types.QueryUnbondingRecordsRequest{IncludeArchived: true}
	allResp, err := s.App.StakedymKeeper.UnbondingRecords(s.Ctx, allReq)
	s.Require().NoError(err, "no error expected when querying all records")

	s.Require().Equal(len(allUnbondingRecords), len(allResp.UnbondingRecords), "all records")
	s.Require().ElementsMatch(allUnbondingRecords, allResp.UnbondingRecords, "all records")
}

func (s *KeeperTestSuite) TestQueryRedemptionRecord() {
	queriedUnbondingRecordId := uint64(2)
	queriedAddress := "address-B"

	unbondingRecords := []types.UnbondingRecord{
		{Id: 1, UnbondingCompletionTimeSeconds: 12345},
		{Id: 2, UnbondingCompletionTimeSeconds: 12346},
		{Id: 3, UnbondingCompletionTimeSeconds: 12347},
		{Id: 4, UnbondingCompletionTimeSeconds: 12348},
	}
	for _, unbondingRecord := range unbondingRecords {
		s.App.StakedymKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}
	// map unbonding record id to unbonding time
	unbondingTimeMap := map[uint64]uint64{}
	for _, unbondingRecord := range unbondingRecords {
		unbondingTimeMap[unbondingRecord.Id] = unbondingRecord.UnbondingCompletionTimeSeconds
	}

	redemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
	}
	for _, redemptionRecord := range redemptionRecords {
		s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordRequest{
		UnbondingRecordId: queriedUnbondingRecordId,
		Address:           queriedAddress,
	}
	resp, err := s.App.StakedymKeeper.RedemptionRecord(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying redemption record")

	s.Require().Equal(queriedUnbondingRecordId, resp.RedemptionRecordResponse.RedemptionRecord.UnbondingRecordId, "redemption record unbonding ID")
	s.Require().Equal(queriedAddress, resp.RedemptionRecordResponse.RedemptionRecord.Redeemer, "redemption record address")
	s.Require().Equal(unbondingTimeMap[queriedUnbondingRecordId], resp.RedemptionRecordResponse.UnbondingCompletionTimeSeconds, "redemption record unbonding time")
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_Address() {
	queriedAddress := "address-B"
	expectedUnbondingRecordIds := []uint64{2, 4}
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
		UnbondingPeriodSeconds: 10000,
	})
	allRedemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
		{UnbondingRecordId: 4, Redeemer: "address-B"},
	}
	for _, redemptionRecord := range allRedemptionRecords {
		s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordsRequest{
		Address: queriedAddress,
	}
	resp, err := s.App.StakedymKeeper.RedemptionRecords(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying redemption records")
	s.Require().Nil(resp.Pagination, "pagination should be nil since it all fits on one page")

	actualUnbondingRecordIds := []uint64{}
	for _, resp := range resp.RedemptionRecordResponses {
		actualUnbondingRecordIds = append(actualUnbondingRecordIds, resp.RedemptionRecord.UnbondingRecordId)
	}
	s.Require().ElementsMatch(expectedUnbondingRecordIds, actualUnbondingRecordIds)
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_UnbondingRecordId() {
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
		UnbondingPeriodSeconds: 10000,
	})
	queriedUnbondingRecordId := uint64(2)
	expectedAddresss := []string{"address-B", "address-D"}
	allRedemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
		{UnbondingRecordId: 2, Redeemer: "address-D"},
	}
	for _, redemptionRecord := range allRedemptionRecords {
		s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordsRequest{
		UnbondingRecordId: queriedUnbondingRecordId,
	}
	resp, err := s.App.StakedymKeeper.RedemptionRecords(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying redemption records")
	s.Require().Nil(resp.Pagination, "pagination should be nil since it all fits on one page")

	actualAddresss := []string{}
	for _, response := range resp.RedemptionRecordResponses {
		actualAddresss = append(actualAddresss, response.RedemptionRecord.Redeemer)
	}
	s.Require().ElementsMatch(expectedAddresss, actualAddresss)
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_Pagination() {
	s.App.StakedymKeeper.SetHostZone(s.Ctx, types.HostZone{
		UnbondingPeriodSeconds: 10000,
	})

	// Set more records than what will fit on one page
	pageLimit := 50
	numExcessRecords := 10
	for i := 0; i < pageLimit+numExcessRecords; i++ {
		s.App.StakedymKeeper.SetRedemptionRecord(s.Ctx, types.RedemptionRecord{
			UnbondingRecordId: uint64(i),
			Redeemer:          fmt.Sprintf("address-%d", i),
		})
	}

	// Query with pagination
	req := &types.QueryRedemptionRecordsRequest{
		Pagination: &query.PageRequest{
			Limit: uint64(pageLimit),
		},
	}
	resp, err := s.App.StakedymKeeper.RedemptionRecords(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying all redemption records")

	// Confirm only the first page was returned
	s.Require().Equal(pageLimit, len(resp.RedemptionRecordResponses), "only the first page should be returned")

	// Attempt one more page, and it should get the remainder
	req = &types.QueryRedemptionRecordsRequest{
		Pagination: &query.PageRequest{
			Key: resp.Pagination.NextKey,
		},
	}
	resp, err = s.App.StakedymKeeper.RedemptionRecords(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying all redemption records on second page")
	s.Require().Equal(numExcessRecords, len(resp.RedemptionRecordResponses), "only the remainder should be returned")
}

func (s *KeeperTestSuite) TestQuerySlashRecords() {
	slashRecords := []types.SlashRecord{
		{Id: 1, Time: 1, NativeAmount: sdkmath.NewInt(1)},
		{Id: 2, Time: 2, NativeAmount: sdkmath.NewInt(2)},
		{Id: 3, Time: 3, NativeAmount: sdkmath.NewInt(3)},
	}
	for _, slashRecord := range slashRecords {
		s.App.StakedymKeeper.SetSlashRecord(s.Ctx, slashRecord)
	}

	req := &types.QuerySlashRecordsRequest{}
	resp, err := s.App.StakedymKeeper.SlashRecords(s.Ctx, req)
	s.Require().NoError(err, "no error expected when querying slash records")
	s.Require().Equal(slashRecords, resp.SlashRecords, "slash records")
}
