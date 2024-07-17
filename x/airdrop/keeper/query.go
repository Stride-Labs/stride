package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

var _ types.QueryServer = Keeper{}

// Queries the configuration for a given airdrop
func (k Keeper) Airdrop(goCtx context.Context, req *types.QueryAirdropRequest) (*types.QueryAirdropResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrop, found := k.GetAirdrop(ctx, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "airdrop %s not found", req.Id)
	}

	return &types.QueryAirdropResponse{Airdrop: &airdrop}, nil
}

// Queries all airdrop configurations
func (k Keeper) AllAirdrops(goCtx context.Context, req *types.QueryAllAirdropsRequest) (*types.QueryAllAirdropsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrops := k.GetAllAirdrops(ctx)
	return &types.QueryAllAirdropsResponse{Airdrops: airdrops}, nil
}

// Queries the allocation for a given user for an airdrop
func (k Keeper) UserAllocation(goCtx context.Context, req *types.QueryUserAllocationRequest) (*types.QueryUserAllocationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	allocations, found := k.GetUserAllocation(ctx, req.AirdropId, req.Address)
	if !found {
		return nil, status.Errorf(codes.NotFound, "allocations not found for airdrop %s and user %s", req.AirdropId, req.Address)
	}

	return &types.QueryUserAllocationResponse{UserAllocation: &allocations}, nil
}

// Queries the allocations for a given user across all airdrops
func (k Keeper) UserAllocations(goCtx context.Context, req *types.QueryUserAllocationsRequest) (*types.QueryUserAllocationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	allocations := k.GetUserAllocationsForAddress(ctx, req.Address)

	return &types.QueryUserAllocationsResponse{UserAllocations: allocations}, nil
}

// Queries all allocations for a given airdrop
func (k Keeper) AllAllocations(goCtx context.Context, req *types.QueryAllAllocationsRequest) (*types.QueryAllAllocationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	allAllocationsStore := prefix.NewStore(store, types.UserAllocationKeyPrefix)
	airdropAllocationsSubstore := prefix.NewStore(allAllocationsStore, types.KeyPrefix(req.AirdropId))

	allocationsOnPage := []types.UserAllocation{}
	pageRes, err := query.Paginate(airdropAllocationsSubstore, req.Pagination, func(key []byte, value []byte) error {
		var userAllocation types.UserAllocation
		if err := k.cdc.Unmarshal(value, &userAllocation); err != nil {
			return err
		}

		allocationsOnPage = append(allocationsOnPage, userAllocation)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAllocationsResponse{
		Allocations: allocationsOnPage,
		Pagination:  pageRes,
	}, nil
}

// Queries the state of an address for an airdrop (daily claim or claim early)
// and the amount claimed and remaining
func (k Keeper) UserSummary(goCtx context.Context, req *types.QueryUserSummaryRequest) (*types.QueryUserSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	airdrop, airdropFound := k.GetAirdrop(ctx, req.AirdropId)
	if !airdropFound {
		return nil, status.Errorf(codes.NotFound, "airdrop %s not found", req.AirdropId)
	}
	allocation, allocationFound := k.GetUserAllocation(ctx, req.AirdropId, req.Address)
	if !allocationFound {
		return nil, status.Errorf(codes.NotFound, "allocations not found for airdrop %s and user %s", req.AirdropId, req.Address)
	}

	windowLength := k.GetParams(ctx).AllocationWindowSeconds
	currentDateIndex, err := airdrop.GetCurrentDateIndex(ctx, windowLength)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	claimType := types.CLAIM_DAILY
	if allocation.Forfeited.GT(sdkmath.ZeroInt()) {
		claimType = types.CLAIM_EARLY
	}

	summary := &types.QueryUserSummaryResponse{
		ClaimType:        claimType.String(),
		Claimed:          allocation.Claimed,
		Claimable:        allocation.GetClaimableAllocation(currentDateIndex),
		Forfeited:        allocation.Forfeited,
		Remaining:        allocation.GetRemainingAllocations(),
		CurrentDateIndex: int64(currentDateIndex),
	}

	return summary, nil
}
