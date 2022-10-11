package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// before removing the validator, set its weight to 0
	_, err := k.ChangeValidatorWeight(goCtx, &types.MsgChangeValidatorWeight{
		Creator:  msg.Creator,
		HostZone: msg.HostZone,
		ValAddr:  msg.ValAddr,
		Weight:   0,
	})
	if err != nil {
		return nil, err
	}

	err = k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)

	if err != nil {
		errMsg := fmt.Sprintf("Validator (%s) not removed from host zone (%s) | err: %s", msg.ValAddr, msg.HostZone, err.Error())
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrapf(types.ErrValidatorNotRemoved, errMsg)
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}
