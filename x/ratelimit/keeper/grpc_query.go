package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Paths(c context.Context, req *types.QueryPathsRequest) (*types.QueryPathsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	paths := k.GetAllPaths(ctx)
	return &types.QueryPathsResponse{Paths: paths}, nil
}

func (k Keeper) Path(c context.Context, req *types.QueryPathRequest) (*types.QueryPathResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	path, found := k.GetPath(ctx, req.Id)
	if !found {
		return &types.QueryPathResponse{}, nil
	}
	return &types.QueryPathResponse{Path: &path}, nil
}

func (k Keeper) Quotas(c context.Context, req *types.QueryQuotasRequest) (*types.QueryQuotasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quotas := k.GetAllQuotas(ctx)
	return &types.QueryQuotasResponse{Quotas: quotas}, nil
}

func (k Keeper) Quota(c context.Context, req *types.QueryQuotaRequest) (*types.QueryQuotaResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	quota, found := k.GetQuota(ctx, req.Name)
	if !found {
		return &types.QueryQuotaResponse{}, nil
	}
	return &types.QueryQuotaResponse{Quota: &quota}, nil
}

func (k Keeper) RateLimits(c context.Context, req *types.QueryRateLimitsRequest) (*types.QueryRateLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimits := k.GetAllRateLimits(ctx)
	return &types.QueryRateLimitsResponse{RateLimits: rateLimits}, nil
}

func (k Keeper) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimit, found := k.GetRateLimit(ctx, req.PathId)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}
