package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

var _ types.QueryServer = Keeper{}

// Query a specific oracle
func (k Keeper) Oracle(c context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	// TODO
	return &types.QueryOracleResponse{}, nil
}

// Query all oracles with s
func (k Keeper) AllOracles(c context.Context, req *types.QueryAllOraclesRequest) (*types.QueryAllOraclesResponse, error) {
	// TODO
	return &types.QueryAllOraclesResponse{}, nil
}

// Query all oracles with a filter on whether they are currently active
func (k Keeper) ActiveOracles(c context.Context, req *types.QueryActiveOraclesRequest) (*types.QueryActiveOraclesResponse, error) {
	// TODO
	return &types.QueryActiveOraclesResponse{}, nil
}

// Query all metrics that currently have an ICA in flight
func (k Keeper) AllPendingMetricUpdates(c context.Context, req *types.QueryAllPendingMetricUpdatesRequest) (*types.QueryAllPendingMetricUpdatesResponse, error) {
	// TODO
	return &types.QueryAllPendingMetricUpdatesResponse{}, nil
}

// Query all metrics that currently have an ICA in flight, with filters
func (k Keeper) PendingMetricUpdates(c context.Context, req *types.QueryPendingMetricUpdatesRequest) (*types.QueryPendingMetricUpdatesResponse, error) {
	// TODO
	return &types.QueryPendingMetricUpdatesResponse{}, nil
}
