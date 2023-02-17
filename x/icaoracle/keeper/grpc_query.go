package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

var _ types.QueryServer = Keeper{}

// Query a specific oracle
func (k Keeper) Oracle(c context.Context, req *types.QueryOracleRequest) (*types.QueryOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	oracle, found := k.GetOracle(ctx, req.ChainId)
	if !found {
		return &types.QueryOracleResponse{}, types.ErrOracleNotFound
	}

	return &types.QueryOracleResponse{Oracle: &oracle}, nil
}

// Query all oracles with s
func (k Keeper) AllOracles(c context.Context, req *types.QueryAllOraclesRequest) (*types.QueryAllOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	oracles := k.GetAllOracles(ctx)
	return &types.QueryAllOraclesResponse{Oracles: oracles}, nil
}

// Query all oracles with a filter on whether they are currently active
func (k Keeper) ActiveOracles(c context.Context, req *types.QueryActiveOraclesRequest) (*types.QueryActiveOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	oracles := []types.Oracle{}
	for _, oracle := range k.GetAllOracles(ctx) {
		if oracle.Active == req.Active {
			oracles = append(oracles, oracle)
		}
	}
	return &types.QueryActiveOraclesResponse{Oracles: oracles}, nil
}

// Query all metrics that currently have an ICA in flight
func (k Keeper) AllPendingMetricUpdates(c context.Context, req *types.QueryAllPendingMetricUpdatesRequest) (*types.QueryAllPendingMetricUpdatesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pendingMetricUpdates := k.GetAllPendingMetricUpdates(ctx)
	return &types.QueryAllPendingMetricUpdatesResponse{PendingUpdates: pendingMetricUpdates}, nil
}

// Query all metrics that currently have an ICA in flight, with filters
func (k Keeper) PendingMetricUpdates(c context.Context, req *types.QueryPendingMetricUpdatesRequest) (*types.QueryPendingMetricUpdatesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pendingMetricUpdates := []types.PendingMetricUpdate{}
	for _, metricUpdate := range k.GetAllPendingMetricUpdates(ctx) {
		if req.MetricKey == "" || req.MetricKey == metricUpdate.Metric.Key {
			if req.OracleChainId == "" || req.OracleChainId == metricUpdate.OracleChainId {
				pendingMetricUpdates = append(pendingMetricUpdates, metricUpdate)
			}
		}
	}

	return &types.QueryPendingMetricUpdatesResponse{PendingUpdates: pendingMetricUpdates}, nil
}
