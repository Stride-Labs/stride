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
	// TODO:
	return &types.MsgAddRateLimitResponse{}, nil
}

func (server msgServer) AddQuota(goCtx context.Context, msg *types.MsgAddQuota) (*types.MsgAddQuotaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	server.Keeper.AddQuota(ctx, types.Quota{
		Name:            msg.Name,
		MaxPercentSend:  msg.MaxPercentSend,
		MaxPercentRecv:  msg.MaxPercentRecv,
		DurationMinutes: msg.DurationMinutes,
		PeriodEnd:       uint64(ctx.BlockTime().Unix()) + msg.DurationMinutes*60,
	})

	return &types.MsgAddQuotaResponse{}, nil
}

func (server msgServer) RemoveQuota(goCtx context.Context, msg *types.MsgRemoveQuota) (*types.MsgRemoveQuotaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	server.Keeper.RemoveQuota(ctx, msg.Name)
	return &types.MsgRemoveQuotaResponse{}, nil
}

func (server msgServer) RemoveRateLimit(goCtx context.Context, msg *types.MsgRemoveRateLimit) (*types.MsgRemoveRateLimitResponse, error) {
	// TODO:
	return &types.MsgRemoveRateLimitResponse{}, nil
}

func (server msgServer) ResetRateLimit(goCtx context.Context, msg *types.MsgResetRateLimit) (*types.MsgResetRateLimitResponse, error) {
	// TODO:
	return &types.MsgResetRateLimitResponse{}, nil
}
