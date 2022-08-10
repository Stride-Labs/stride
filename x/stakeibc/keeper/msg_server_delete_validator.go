package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validatorRemoved := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if !validatorRemoved {
		k.Logger(ctx).Error(fmt.Sprintf("Validator %s not removed from the host zone %s", msg.ValAddr, msg.HostZone))
		return nil, types.ErrValidatorNotRemoved
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}
