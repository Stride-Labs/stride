package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetNumValidators(goCtx context.Context, msg *types.MsgSetNumValidators) (*types.MsgSetNumValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx
	// k.stakingKeeper.SetNumValidators(ctx, msg.NumValidators)

	return &types.MsgSetNumValidatorsResponse{}, nil
}
