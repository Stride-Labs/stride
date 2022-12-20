package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

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

func (k Keeper) RateLimitByChainId(c context.Context, req *types.QueryRateLimitsByChainIdRequest) (*types.QueryRateLimitsByChainIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {

		// Determine the client state from the channel Id
		_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, rateLimit.Path.ChannelId)
		if err != nil {
			return &types.QueryRateLimitsByChainIdResponse{}, sdkerrors.Wrapf(types.ErrInvalidClientState, "Unable to fetch client state from channelId")
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if !ok {
			return &types.QueryRateLimitsByChainIdResponse{}, sdkerrors.Wrapf(types.ErrInvalidClientState, "Client state is not tendermint")
		}

		// If the chain ID matches, add the rate limit to the returned list
		if client.ChainId == req.ChainId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChainIdResponse{RateLimits: rateLimits}, nil
}
