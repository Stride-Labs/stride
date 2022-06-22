package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgDeleteValidatorResponse{}, nil
}
