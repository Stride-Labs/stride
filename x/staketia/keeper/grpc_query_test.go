package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) TestQueryHostZone() {
	chainId := "chain-0"
	hostZone := types.HostZone{
		ChainId: chainId,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	req := &types.QueryHostZoneRequest{}
	resp, err := s.App.StaketiaKeeper.HostZone(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying host zone")
	s.Require().Equal(chainId, resp.HostZone.ChainId, "host zone chain-id from query")
}

func (s *KeeperTestSuite) TestQueryDelegationRecords() {
	// Create active delegation records
	activeDelegationRecords := s.addDelegationRecords()

	// Create an archived version of each of the above records by removing
	// the record (i.e. archiving it) and then recreating it
	archivedDelegationRecords := []types.DelegationRecord{}
	for _, delegationRecord := range activeDelegationRecords {
		// Archive (which removes from the active store, and writes to the archive store)
		s.App.StaketiaKeeper.RemoveDelegationRecord(s.Ctx, delegationRecord.Id)

		// Set the original record back to the active store
		s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, delegationRecord)

		archivedRecord := delegationRecord
		archivedRecord.Status = types.DELEGATION_ARCHIVE
		archivedDelegationRecords = append(archivedDelegationRecords, archivedRecord)
	}
	allDelegationRecords := append(activeDelegationRecords, archivedDelegationRecords...)

	// Test a query with no archived records
	activeReq := &types.QueryDelegationRecordsRequest{IncludeArchived: false}
	activeResp, err := s.App.StaketiaKeeper.DelegationRecords(sdk.WrapSDKContext(s.Ctx), activeReq)
	s.Require().NoError(err, "no error expected when querying active records")

	s.Require().Equal(len(activeDelegationRecords), len(activeResp.DelegationRecords), "number of active records")
	s.Require().ElementsMatch(activeDelegationRecords, activeResp.DelegationRecords, "active records")

	// Test a query with all records (including archived records)
	allReq := &types.QueryDelegationRecordsRequest{IncludeArchived: true}
	allResp, err := s.App.StaketiaKeeper.DelegationRecords(sdk.WrapSDKContext(s.Ctx), allReq)
	s.Require().NoError(err, "no error expected when querying all records")

	s.Require().Equal(len(allDelegationRecords), len(allResp.DelegationRecords), "all records")
	s.Require().ElementsMatch(allDelegationRecords, allResp.DelegationRecords, "all records")
}

func (s *KeeperTestSuite) TestQueryUnbondingRecords() {
	// Create active unbondin records
	activeUnbondingRecords := s.addUnbondingRecords()

	// Create an archived version of each of the above records by removing
	// the record (i.e. archiving it) and then recreating it
	archivedUnbondingRecords := []types.UnbondingRecord{}
	for _, unbondingRecord := range activeUnbondingRecords {
		// Archive (which removes from the active store, and writes to the archive store)
		s.App.StaketiaKeeper.RemoveUnbondingRecord(s.Ctx, unbondingRecord.Id)

		// Set the original record back to the active store
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)

		archivedRecord := unbondingRecord
		archivedRecord.Status = types.UNBONDING_ARCHIVE
		archivedUnbondingRecords = append(archivedUnbondingRecords, archivedRecord)
	}
	allUnbondingRecords := append(activeUnbondingRecords, archivedUnbondingRecords...)

	// Test a query with no archived records
	activeReq := &types.QueryUnbondingRecordsRequest{IncludeArchived: false}
	activeResp, err := s.App.StaketiaKeeper.UnbondingRecords(sdk.WrapSDKContext(s.Ctx), activeReq)
	s.Require().NoError(err, "no error expected when querying active records")

	s.Require().Equal(len(activeUnbondingRecords), len(activeResp.UnbondingRecords), "number of active records")
	s.Require().ElementsMatch(activeUnbondingRecords, activeResp.UnbondingRecords, "active records")

	// Test a query with no all records
	allReq := &types.QueryUnbondingRecordsRequest{IncludeArchived: true}
	allResp, err := s.App.StaketiaKeeper.UnbondingRecords(sdk.WrapSDKContext(s.Ctx), allReq)
	s.Require().NoError(err, "no error expected when querying all records")

	s.Require().Equal(len(allUnbondingRecords), len(allResp.UnbondingRecords), "all records")
	s.Require().ElementsMatch(allUnbondingRecords, allResp.UnbondingRecords, "all records")
}

func (s *KeeperTestSuite) TestQueryRedemptionRecord() {
	queriedUnbondingRecordId := uint64(2)
	queriedAddress := "address-B"

	redemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
	}
	for _, redemptionRecord := range redemptionRecords {
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordRequest{
		UnbondingRecordId: queriedUnbondingRecordId,
		Address:           queriedAddress,
	}
	resp, err := s.App.StaketiaKeeper.RedemptionRecord(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying redemption record")

	s.Require().Equal(queriedUnbondingRecordId, resp.RedemptionRecord.UnbondingRecordId, "redemption record unbonding ID")
	s.Require().Equal(queriedAddress, resp.RedemptionRecord.Redeemer, "redemption record address")
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_Address() {
	queriedAddress := "address-B"
	expectedUnbondingRecordIds := []uint64{2, 4}
	allRedemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
		{UnbondingRecordId: 4, Redeemer: "address-B"},
	}
	for _, redemptionRecord := range allRedemptionRecords {
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordsRequest{
		Address: queriedAddress,
	}
	resp, err := s.App.StaketiaKeeper.RedemptionRecords(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying redemption records")
	s.Require().Nil(resp.Pagination, "pagination should be nil since it all fits on one page")

	actualUnbondingRecordIds := []uint64{}
	for _, record := range resp.RedemptionRecords {
		actualUnbondingRecordIds = append(actualUnbondingRecordIds, record.UnbondingRecordId)
	}
	s.Require().ElementsMatch(expectedUnbondingRecordIds, actualUnbondingRecordIds)
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_UnbondingRecordId() {
	queriedUnbondingRecordId := uint64(2)
	expectedAddresss := []string{"address-B", "address-D"}
	allRedemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, Redeemer: "address-A"},
		{UnbondingRecordId: 2, Redeemer: "address-B"},
		{UnbondingRecordId: 3, Redeemer: "address-C"},
		{UnbondingRecordId: 2, Redeemer: "address-D"},
	}
	for _, redemptionRecord := range allRedemptionRecords {
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	req := &types.QueryRedemptionRecordsRequest{
		UnbondingRecordId: queriedUnbondingRecordId,
	}
	resp, err := s.App.StaketiaKeeper.RedemptionRecords(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying redemption records")
	s.Require().Nil(resp.Pagination, "pagination should be nil since it all fits on one page")

	actualAddresss := []string{}
	for _, record := range resp.RedemptionRecords {
		actualAddresss = append(actualAddresss, record.Redeemer)
	}
	s.Require().ElementsMatch(expectedAddresss, actualAddresss)
}

func (s *KeeperTestSuite) TestQueryAllRedemptionRecords_Pagination() {
	// Set more records than what will fit on one page
	pageLimit := 50
	numExcessRecords := 10
	for i := 0; i < pageLimit+numExcessRecords; i++ {
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, types.RedemptionRecord{
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
	resp, err := s.App.StaketiaKeeper.RedemptionRecords(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all redemption records")

	// Confirm only the first page was returned
	s.Require().Equal(pageLimit, len(resp.RedemptionRecords), "only the first page should be returned")

	// Attempt one more page, and it should get the remainder
	req = &types.QueryRedemptionRecordsRequest{
		Pagination: &query.PageRequest{
			Key: resp.Pagination.NextKey,
		},
	}
	resp, err = s.App.StaketiaKeeper.RedemptionRecords(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all redemption records on second page")
	s.Require().Equal(numExcessRecords, len(resp.RedemptionRecords), "only the remainder should be returned")
}

func (s *KeeperTestSuite) TestQuerySlashRecords() {
	slashRecords := []types.SlashRecord{
		{Id: 1, Time: 1, SttokenAmount: sdkmath.NewInt(1)},
		{Id: 2, Time: 2, SttokenAmount: sdkmath.NewInt(2)},
		{Id: 3, Time: 3, SttokenAmount: sdkmath.NewInt(3)},
	}
	for _, slashRecord := range slashRecords {
		s.App.StaketiaKeeper.SetSlashRecord(s.Ctx, slashRecord)
	}

	req := &types.QuerySlashRecordsRequest{}
	resp, err := s.App.StaketiaKeeper.SlashRecords(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying slash records")
	s.Require().Equal(slashRecords, resp.SlashRecords, "slash records")
}
