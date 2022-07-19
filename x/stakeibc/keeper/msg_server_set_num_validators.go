package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetNumValidators(goCtx context.Context, msg *types.MsgSetNumValidators) (*types.MsgSetNumValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ps := k.StakingKeeper.GetParams(ctx)
	ps.MaxValidators = msg.NumValidators
	k.StakingKeeper.SetParams(ctx, ps)

	return &types.MsgSetNumValidatorsResponse{}, nil
}
