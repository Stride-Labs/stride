package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.AddValidatorToHostZone(ctx, msg, false)
	if err != nil {
		return nil, err
	}
	return &types.MsgAddValidatorResponse{}, nil
}
