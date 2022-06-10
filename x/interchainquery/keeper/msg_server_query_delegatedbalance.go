package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) QueryDelegatedbalance(goCtx context.Context, msg *types.MsgQueryDelegatedbalance) (*types.MsgQueryDelegatedbalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgQueryDelegatedbalanceResponse{}, nil
}
