package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Stride-Labs/stride/v25/x/airdrop/types"
)

func (s *KeeperTestSuite) TestQueryAirdrop() {
	// Create mutiple airdrops and then query for a specific one
	airdrops := s.addAirdrops()
	expectedAirdrop := airdrops[1]

	// Update the date boundaries
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)      // 10 days later
	clawbackDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC) // 10 more days later

	expectedAirdropLength := int64(10)
	expectedAirdrop.DistributionStartDate = &startTime
	expectedAirdrop.DistributionEndDate = &endTime
	expectedAirdrop.ClawbackDate = &clawbackDate
	s.App.AirdropKeeper.SetAirdrop(s.Ctx, expectedAirdrop)

	// Set the block time so that we're on the third day
	blockTime := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	expectedDateIndex := int64(2) // third day is index 2
	s.Ctx = s.Ctx.WithBlockTime(blockTime)

	// Query for the airdrop
	req := &types.QueryAirdropRequest{
		Id: expectedAirdrop.Id,
	}
	respAirdrop, err := s.App.AirdropKeeper.Airdrop(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying an airdrop")

	// Confirm all the airdrop fields
	s.Require().Equal(expectedAirdrop.Id, respAirdrop.Id, "airdrop id")
	s.Require().Equal(expectedAirdrop.RewardDenom, respAirdrop.RewardDenom, "airdrop reward denom")

	s.Require().Equal(expectedAirdrop.DistributionStartDate, respAirdrop.DistributionStartDate, "airdrop start")
	s.Require().Equal(expectedAirdrop.DistributionEndDate, respAirdrop.DistributionEndDate, "airdrop end")
	s.Require().Equal(expectedAirdrop.ClawbackDate, respAirdrop.ClawbackDate, "airdrop clawback")
	s.Require().Equal(expectedAirdrop.ClaimTypeDeadlineDate, respAirdrop.ClaimTypeDeadlineDate, "airdrop deadline")

	s.Require().Equal(expectedAirdrop.EarlyClaimPenalty, respAirdrop.EarlyClaimPenalty, "airdrop penalty")
	s.Require().Equal(expectedAirdrop.DistributorAddress, respAirdrop.DistributorAddress, "airdrop distributor")
	s.Require().Equal(expectedAirdrop.AllocatorAddress, respAirdrop.AllocatorAddress, "airdrop allocator")
	s.Require().Equal(expectedAirdrop.LinkerAddress, respAirdrop.LinkerAddress, "airdrop linker")

	s.Require().Equal(expectedDateIndex, respAirdrop.CurrentDateIndex, "airdrop date index")
	s.Require().Equal(expectedAirdropLength, respAirdrop.AirdropLength, "airdrop length")

	// Update the block time so the airdrop hasn't started yet
	// Confirm the current date index is -1
	s.Ctx = s.Ctx.WithBlockTime(startTime.Add(-1 * time.Hour))
	respAirdrop, err = s.App.AirdropKeeper.Airdrop(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying an airdrop before it has started")
	s.Require().Equal(int64(-1), respAirdrop.CurrentDateIndex, "date index before airdrop")

	// Update the block time so the airdrop distribution has ended, but the clawback data hasn't hit
	// Confirm the current date index is 9 (last index of the 10 day array)
	s.Ctx = s.Ctx.WithBlockTime(clawbackDate.Add(-1 * time.Hour))
	respAirdrop, err = s.App.AirdropKeeper.Airdrop(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying an airdrop after distribution ended")
	s.Require().Equal(int64(9), respAirdrop.CurrentDateIndex, "date index after distribution")

	// Update the block time so the clawback date has passed
	// Confirm the current date index is -1
	s.Ctx = s.Ctx.WithBlockTime(clawbackDate.Add(time.Hour))
	respAirdrop, err = s.App.AirdropKeeper.Airdrop(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying an airdrop after the clawback date")
	s.Require().Equal(int64(-1), respAirdrop.CurrentDateIndex, "date index after clawback")
}

func (s *KeeperTestSuite) TestQueryAllAirdrops() {
	// Create mulitple airdrops and then query for all of them
	expectedAirdrops := s.addAirdrops()

	req := &types.QueryAllAirdropsRequest{}
	resp, err := s.App.AirdropKeeper.AllAirdrops(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all airdrops")
	s.Require().Equal(expectedAirdrops, resp.Airdrops, "airdrops")
}

func (s *KeeperTestSuite) TestQueryAllAllocations() {
	// Create allocations for a give airdrop Id
	expectedAllocations := []types.UserAllocation{}
	for i := 1; i <= 5; i++ {
		userAllocation := newUserAllocation(AirdropId, fmt.Sprintf("address-%d", i))
		s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)
		expectedAllocations = append(expectedAllocations, userAllocation)
	}

	// Query for that airdrop ID and confirm all were returned
	req := &types.QueryAllAllocationsRequest{
		AirdropId: AirdropId,
	}
	resp, err := s.App.AirdropKeeper.AllAllocations(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all allocations")
	s.Require().Equal(expectedAllocations, resp.Allocations)
}

func (s *KeeperTestSuite) TestQueryAllAllocations_Pagination() {
	// Set more allocations than what will fit on one page
	pageLimit := 50
	numExcessRecords := 10
	for i := 0; i < pageLimit+numExcessRecords; i++ {
		s.App.AirdropKeeper.SetUserAllocation(s.Ctx, types.UserAllocation{
			AirdropId: fmt.Sprintf("airdrop-%d", i),
			Address:   fmt.Sprintf("address-%d", i),
		})
	}

	// Query with pagination
	req := &types.QueryAllAllocationsRequest{
		Pagination: &query.PageRequest{
			Limit: uint64(pageLimit),
		},
	}
	resp, err := s.App.AirdropKeeper.AllAllocations(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all allocations")

	// Confirm only the first page was returned
	s.Require().Equal(pageLimit, len(resp.Allocations), "only the first page should be returned")

	// Attempt one more page, and it should get the remainder
	req = &types.QueryAllAllocationsRequest{
		Pagination: &query.PageRequest{
			Key: resp.Pagination.NextKey,
		},
	}
	resp, err = s.App.AirdropKeeper.AllAllocations(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all allocations on second page")
	s.Require().Equal(numExcessRecords, len(resp.Allocations), "only the remainder should be returned")
}

func (s *KeeperTestSuite) TestQueryUserAllocation() {
	// Create allocations across different airdrops/addresses and then query for one of them
	allocations := s.addUserAllocations()
	expectedAllocation := allocations[1]

	req := &types.QueryUserAllocationRequest{
		AirdropId: expectedAllocation.AirdropId,
		Address:   expectedAllocation.Address,
	}
	resp, err := s.App.AirdropKeeper.UserAllocation(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying a user allocation")
	s.Require().Equal(expectedAllocation, *resp.UserAllocation, "user allocation")
}

func (s *KeeperTestSuite) TestQueryUserAllocations() {
	// Create allocations for a given address
	expectedAllocations := []types.UserAllocation{}
	for i := 1; i <= 5; i++ {
		airdropId := fmt.Sprintf("airdrop-%d", i)
		s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{Id: airdropId})

		userAllocation := newUserAllocation(airdropId, UserAddress)
		s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)
		expectedAllocations = append(expectedAllocations, userAllocation)
	}

	// Query the allocations for that address
	req := &types.QueryUserAllocationsRequest{
		Address: UserAddress,
	}
	resp, err := s.App.AirdropKeeper.UserAllocations(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user allocations")
	s.Require().Equal(expectedAllocations, resp.UserAllocations, "user allocations")
}

func (s *KeeperTestSuite) TestQueryUserSummary() {
	// Create a user allocation with 10 claimed and a few days remaining
	claimed := sdkmath.NewInt(10)
	forfeited := sdkmath.ZeroInt()
	remaining := sdkmath.NewInt(1 + 5 + 3)
	claimable := sdkmath.NewInt(1 + 5)

	userAllocation := types.UserAllocation{
		AirdropId: AirdropId,
		Address:   UserAddress,
		Claimed:   claimed,
		Forfeited: forfeited,
		Allocations: []sdkmath.Int{
			sdkmath.ZeroInt(),
			sdkmath.NewInt(1),
			sdkmath.NewInt(5), // today
			sdkmath.NewInt(3),
		},
	}
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)
	s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
		Id:                    AirdropId,
		DistributionStartDate: &DistributionStartDate,
		DistributionEndDate:   &DistributionEndDate,
		ClawbackDate:          &ClawbackDate,
	})

	s.Ctx = s.Ctx.WithBlockTime(DistributionStartDate.Add(time.Hour * 49))

	// Query the summary and confirm the remaining total is correct
	req := &types.QueryUserSummaryRequest{
		AirdropId: AirdropId,
		Address:   UserAddress,
	}
	resp, err := s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary")
	s.Require().Equal(claimed, resp.Claimed, "amount claimed")
	s.Require().Equal(forfeited, resp.Forfeited, "amount forfeited")
	s.Require().Equal(remaining, resp.Remaining, "amount remaining")
	s.Require().Equal(claimable, resp.Claimable, "amount claimable")
	s.Require().Equal(types.CLAIM_DAILY.String(), resp.ClaimType, "claim type")

	// Update the user so that it appears they claimed early and confirm the type change
	userAllocation.Forfeited = sdkmath.OneInt()
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary again")
	s.Require().Equal(types.CLAIM_EARLY.String(), resp.ClaimType, "claim type")

	// Update the block time so it appears as if the airdrop has not started
	// Then check that claimable is 0
	s.Ctx = s.Ctx.WithBlockTime(DistributionStartDate.Add(-1 * time.Hour))

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary before airdrop")
	s.Require().Equal(remaining, resp.Remaining, "remaining")
	s.Require().Equal(int64(0), resp.Claimable.Int64(), "claimable before airdrop")

	// Update the block time so it appears as if the airdrop has ended
	// Then check that claimable is 0
	s.Ctx = s.Ctx.WithBlockTime(ClawbackDate.Add(time.Hour))

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary after airdrop")
	s.Require().Equal(int64(0), resp.Claimable.Int64(), "claimable after airdrop")
}
