package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgAddValidatorResponse{}, nil
}
