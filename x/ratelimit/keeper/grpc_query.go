package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

var _ types.QueryServer = Keeper{}

// Query all rate limits
func (k Keeper) AllRateLimits(c context.Context, req *types.QueryAllRateLimitsRequest) (*types.QueryAllRateLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimits := k.GetAllRateLimits(ctx)
	return &types.QueryAllRateLimitsResponse{RateLimits: rateLimits}, nil
}

// Query a rate limit by denom and channelId
func (k Keeper) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimit, found := k.GetRateLimit(ctx, req.Denom, req.ChannelId)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}

// Query all rate limits for a given chain
func (k Keeper) RateLimitsByChainId(c context.Context, req *types.QueryRateLimitsByChainIdRequest) (*types.QueryRateLimitsByChainIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {

		// Determine the client state from the channel Id
		_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, rateLimit.Path.ChannelId)
		if err != nil {
			return &types.QueryRateLimitsByChainIdResponse{}, errorsmod.Wrapf(types.ErrInvalidClientState, "Unable to fetch client state from channelId")
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if !ok {
			return &types.QueryRateLimitsByChainIdResponse{}, errorsmod.Wrapf(types.ErrInvalidClientState, "Client state is not tendermint")
		}

		// If the chain ID matches, add the rate limit to the returned list
		if client.ChainId == req.ChainId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChainIdResponse{RateLimits: rateLimits}, nil
}

// Query all rate limits for a given channel
func (k Keeper) RateLimitsByChannelId(c context.Context, req *types.QueryRateLimitsByChannelIdRequest) (*types.QueryRateLimitsByChannelIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		// If the channel ID matches, add the rate limit to the returned list
		if rateLimit.Path.ChannelId == req.ChannelId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChannelIdResponse{RateLimits: rateLimits}, nil
}
