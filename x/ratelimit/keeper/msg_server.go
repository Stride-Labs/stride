package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v3/x/ratelimit/types"
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
	// TODO:
	return &types.MsgAddQuotaResponse{}, nil
}

func (server msgServer) ResetRateLimit(goCtx context.Context, msg *types.MsgResetRateLimit) (*types.MsgResetRateLimitResponse, error) {
	// TODO:
	return &types.MsgResetRateLimitResponse{}, nil
}

func (server msgServer) RemoveRateLimit(goCtx context.Context, msg *types.MsgRemoveRateLimit) (*types.MsgRemoveRateLimitResponse, error) {
	// TODO:
	return &types.MsgRemoveRateLimitResponse{}, nil
}
