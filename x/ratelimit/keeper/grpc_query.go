package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RateLimits(c context.Context, req *types.QueryRateLimitsRequest) (*types.QueryRateLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimits := k.GetAllRateLimits(ctx)
	return &types.QueryRateLimitsResponse{RateLimits: rateLimits}, nil
}

func (k Keeper) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimit, found := k.GetRateLimit(ctx, req.Denom, req.ChannelId)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}
