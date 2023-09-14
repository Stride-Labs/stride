package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
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

// Query metrics with optional filters
func (k Keeper) Metrics(c context.Context, req *types.QueryMetricsRequest) (*types.QueryMetricsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	metrics := []types.Metric{}
	for _, metric := range k.GetAllMetrics(ctx) {
		metricKeyMatch := req.MetricKey == "" || req.MetricKey == metric.Key
		metricOracleMatch := req.OracleChainId == "" || req.OracleChainId == metric.DestinationOracle

		if metricKeyMatch && metricOracleMatch {
			metrics = append(metrics, metric)
		}
	}

	return &types.QueryMetricsResponse{Metrics: metrics}, nil
}
