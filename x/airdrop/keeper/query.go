package keeper

import (
	"context"

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
	allocationsStore := prefix.NewStore(store, types.UserAllocationKeyPrefix)

	allocationsOnPage := []types.UserAllocation{}
	pageRes, err := query.Paginate(allocationsStore, req.Pagination, func(key []byte, value []byte) error {
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

	allocation, found := k.GetUserAllocation(ctx, req.AirdropId, req.Address)
	if !found {
		return nil, status.Errorf(codes.NotFound, "allocations not found for airdrop %s and user %s", req.AirdropId, req.Address)
	}

	summary := &types.QueryUserSummaryResponse{
		ClaimType: allocation.ClaimType.String(),
		Claimed:   allocation.Claimed,
		Forfeited: allocation.Forfeited,
		Remaining: allocation.RemainingAllocations(),
	}

	return summary, nil
}
