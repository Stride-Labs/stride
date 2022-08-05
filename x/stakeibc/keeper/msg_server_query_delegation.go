package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) QueryDelegation(goCtx context.Context, msg *types.MsgQueryDelegation) (*types.MsgQueryDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgQueryDelegationResponse{}, nil
}
