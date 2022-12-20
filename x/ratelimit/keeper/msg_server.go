package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Adds a new rate limit. Fails if the rate limit already exists or the channel value is 0
func (server msgServer) AddRateLimit(goCtx context.Context, msg *types.MsgAddRateLimit) (*types.MsgAddRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	path := types.Path{
		Denom:     msg.Denom,
		ChannelId: msg.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: msg.MaxPercentSend,
		MaxPercentRecv: msg.MaxPercentRecv,
		DurationHours:  msg.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdk.ZeroInt(),
		Outflow:      sdk.ZeroInt(),
		ChannelValue: server.Keeper.GetChannelValue(ctx, msg.Denom),
	}
	if flow.ChannelValue.IsZero() {
		return nil, types.ErrZeroChannelValue
	}

	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if found {
		return nil, types.ErrRateLimitKeyAlreadyExists
	}

	server.Keeper.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return &types.MsgAddRateLimitResponse{}, nil
}

// Updates an existing rate limit. Fails if the rate limit doesn't exist
func (server msgServer) UpdateRateLimit(goCtx context.Context, msg *types.MsgUpdateRateLimit) (*types.MsgUpdateRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	path := types.Path{
		Denom:     msg.Denom,
		ChannelId: msg.ChannelId,
	}
	quota := types.Quota{
		MaxPercentSend: msg.MaxPercentSend,
		MaxPercentRecv: msg.MaxPercentRecv,
		DurationHours:  msg.DurationHours,
	}
	flow := types.Flow{
		Inflow:       sdk.ZeroInt(),
		Outflow:      sdk.ZeroInt(),
		ChannelValue: server.Keeper.GetChannelValue(ctx, msg.Denom),
	}

	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitKeyNotFound
	}

	server.Keeper.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return &types.MsgUpdateRateLimitResponse{}, nil
}

// Removes a rate limit. Fails if the rate limit doesn't exist
func (server msgServer) RemoveRateLimit(goCtx context.Context, msg *types.MsgRemoveRateLimit) (*types.MsgRemoveRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitKeyNotFound
	}

	server.Keeper.RemoveRateLimit(ctx, msg.Denom, msg.ChannelId)
	return &types.MsgRemoveRateLimitResponse{}, nil
}

// Resets the flow on a rate limit. Fails if the rate limit doesn't exist
func (server msgServer) ResetRateLimit(goCtx context.Context, msg *types.MsgResetRateLimit) (*types.MsgResetRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := server.Keeper.ResetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if err != nil {
		return &types.MsgResetRateLimitResponse{}, err
	}
	return &types.MsgResetRateLimitResponse{}, nil
}
