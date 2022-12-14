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
		Inflow:       0,
		Outflow:      0,
		ChannelValue: server.Keeper.GetChannelValue(ctx, msg.Denom).Uint64(),
	}

	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if found {
		return nil, types.ErrRateLimitKeyDuplicated
	}

	server.Keeper.SetRateLimit(ctx, types.RateLimit{
		Path:  &path,
		Quota: &quota,
		Flow:  &flow,
	})

	return &types.MsgAddRateLimitResponse{}, nil
}

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
		Inflow:       0,
		Outflow:      0,
		ChannelValue: server.Keeper.GetChannelValue(ctx, msg.Denom).Uint64(),
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

func (server msgServer) RemoveRateLimit(goCtx context.Context, msg *types.MsgRemoveRateLimit) (*types.MsgRemoveRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitKeyNotFound
	}

	server.Keeper.RemoveRateLimit(ctx, msg.Denom, msg.ChannelId)
	return &types.MsgRemoveRateLimitResponse{}, nil
}

func (server msgServer) ResetRateLimit(goCtx context.Context, msg *types.MsgResetRateLimit) (*types.MsgResetRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rateLimit, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitKeyNotFound
	}

	rateLimit.Flow.Inflow = 0
	rateLimit.Flow.Outflow = 0

	server.Keeper.SetRateLimit(ctx, rateLimit)
	return &types.MsgResetRateLimitResponse{}, nil
}
