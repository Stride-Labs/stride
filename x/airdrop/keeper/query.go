package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

var _ types.QueryServer = Keeper{}

// Queries all allocations across all addresses
func (k Keeper) AllAllocations(ctx context.Context, req *types.QueryAllAllocationsRequest) (*types.QueryAllAllocationsResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}

// Queries the allocation for a given user for an airdrop
func (k Keeper) UserAllocation(ctx context.Context, req *types.QueryUserAllocationRequest) (*types.QueryUserAllocationResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}

// Queries the allocations for a given user across all airdrops
func (k Keeper) UserAllocations(ctx context.Context, req *types.QueryUserAllocationsRequest) (*types.QueryUserAllocationsResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}

// Queries the state of an address for an airdrop (daily claim, claim & stake,
// upfront)
func (k Keeper) UserClaimType(ctx context.Context, req *types.QueryUserClaimTypeRequest) (*types.QueryUserClaimTypeResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}

// Queries all airdrop configurations
func (k Keeper) AllAirdrops(ctx context.Context, req *types.QueryAirdropsRequest) (*types.QueryAirdropsResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}

// Queries the configuration for a given airdrop
func (k Keeper) Airdrop(ctx context.Context, req *types.QueryAirdropRequest) (*types.QueryAirdropResponse, error) {
	// TODO[airdrop] implement logic
	return nil, nil
}
