package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if err != nil {
		errMsg := fmt.Sprintf("Validator (%s) not removed from host zone (%s) | err: %s", msg.ValAddr, msg.HostZone, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrValidatorNotRemoved, errMsg)
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}
