package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorRemoved := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if !validatorRemoved {
		k.Logger(ctx).Error(fmt.Sprintf("Validator %s not found in host zone %s", msg.ValAddr, msg.HostZone))
		return nil, types.ErrValidatorNotFound
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}
