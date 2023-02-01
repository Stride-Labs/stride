package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

func (k msgServer) LiquidVote(goCtx context.Context, msg *types.MsgLiquidVote) (*types.MsgLiquidVoteResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.MsgLiquidVoteResponse{}, nil
}
