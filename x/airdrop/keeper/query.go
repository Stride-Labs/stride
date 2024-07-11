package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) AllAllocations(ctx context.Context, req *types.QueryAllAllocationsRequest) (*types.QueryAllAllocationsResponse, error) {
	// TODO implement logic
	return nil, nil
}

func (k Keeper) UserAllocation(ctx context.Context, req *types.QueryUserAllocationRequest) (*types.QueryUserAllocationResponse, error) {
	// TODO implement logic
	return nil, nil
}

func (k Keeper) UserAllocations(ctx context.Context, req *types.QueryUserAllocationsRequest) (*types.QueryUserAllocationsResponse, error) {
	// TODO implement logic
	return nil, nil
}

func (k Keeper) UserClaimType(ctx context.Context, req *types.QueryUserClaimTypeRequest) (*types.QueryUserClaimTypeResponse, error) {
	// TODO implement logic
	return nil, nil
}

func (k Keeper) Airdrops(ctx context.Context, req *types.QueryAirdropsRequest) (*types.QueryAirdropsResponse, error) {
	// TODO implement logic
	return nil, nil
}

func (k Keeper) Airdrop(ctx context.Context, req *types.QueryAirdropRequest) (*types.QueryAirdropResponse, error) {
	// TODO implement logic
	return nil, nil
}
