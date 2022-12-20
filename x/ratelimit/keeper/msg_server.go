package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

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

	// Confirm the channel value is not zero
	channelValue := server.Keeper.GetChannelValue(ctx, msg.Denom)
	if channelValue.IsZero() {
		return nil, types.ErrZeroChannelValue
	}

	// Confirm the rate limit does not already exist
	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if found {
		return nil, types.ErrRateLimitKeyAlreadyExists
	}

	// Confirm the channel exists
	_, found = server.Keeper.channelKeeper.GetChannel(ctx, transfertypes.PortID, msg.ChannelId)
	if !found {
		return nil, types.ErrChannelNotFound
	}

	// Create and store the rate limit object
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
		ChannelValue: channelValue,
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

	// Confirm the rate limit exists
	_, found := server.Keeper.GetRateLimit(ctx, msg.Denom, msg.ChannelId)
	if !found {
		return nil, types.ErrRateLimitKeyNotFound
	}

	// Update the rate limit object with the new quota information
	// The flow should also get reset to 0
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
