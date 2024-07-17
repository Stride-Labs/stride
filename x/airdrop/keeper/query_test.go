package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func (s *KeeperTestSuite) TestQueryAirdrop() {
	// Create mutiple airdrops and then query for a specific one
	airdrops := s.addAirdrops()
	expectedAirdrop := airdrops[1]

	req := &types.QueryAirdropRequest{
		Id: expectedAirdrop.Id,
	}
	resp, err := s.App.AirdropKeeper.Airdrop(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying an airdrop")
	s.Require().Equal(expectedAirdrop, *resp.Airdrop, "airdrop")
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
	dateIndex := int64(2)

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
	s.Require().Equal(dateIndex, resp.CurrentDateIndex, "todays index")

	// Update the user so that it appears they claimed early and confirm the type change
	userAllocation.Forfeited = sdkmath.OneInt()
	s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary again")
	s.Require().Equal(types.CLAIM_EARLY.String(), resp.ClaimType, "claim type")

	// Update the block time so it appears as if the airdrop has not started
	// Then check that the date index is -1 and claimable is 0
	s.Ctx = s.Ctx.WithBlockTime(DistributionStartDate.Add(-1 * time.Hour))

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary before airdrop")
	s.Require().Equal(int64(-1), resp.CurrentDateIndex, "date index")
	s.Require().Equal(remaining, resp.Remaining, "date index")
	s.Require().Equal(int64(0), resp.Claimable.Int64(), "date index")

	// Update the block time so it appears as if the airdrop has not started
	// Then check that the date index is -1 and claimable is 0
	s.Ctx = s.Ctx.WithBlockTime(DistributionStartDate.Add(-1 * time.Hour))

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary before airdrop")
	s.Require().Equal(int64(-1), resp.CurrentDateIndex, "date index")
	s.Require().Equal(int64(0), resp.Claimable.Int64(), "date index")

	// Update the block time so it appears as if the airdrop has ended
	// Then check that the date index is -1 and claimable is 0
	s.Ctx = s.Ctx.WithBlockTime(ClawbackDate.Add(time.Hour))

	resp, err = s.App.AirdropKeeper.UserSummary(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying user summary before airdrop")
	s.Require().Equal(int64(-1), resp.CurrentDateIndex, "date index")
	s.Require().Equal(int64(0), resp.Claimable.Int64(), "date index")
}
