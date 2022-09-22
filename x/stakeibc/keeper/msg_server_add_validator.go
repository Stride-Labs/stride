package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	// TODO TEST-129 restrict this to governance module. add gov module whitelist hooks more broadly
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		errMsg := fmt.Sprintf("Host Zone (%s) not found", msg.HostZone)
		k.Logger(ctx).Error(errMsg)
		return nil, sdkerrors.Wrap(types.ErrHostZoneNotFound, errMsg)
	}
	validators := hostZone.Validators
	// check that we don't already have this validator
	for _, validator := range validators {
		if validator.GetAddress() == msg.Address {
			errMsg := fmt.Sprintf("Validator address (%s) already exists on Host Zone (%s)", msg.Address, msg.HostZone)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrap(types.ErrValidatorAlreadyExists, errMsg)
		}
		if validator.Name == msg.Name {
			errMsg := fmt.Sprintf("Validator name (%s) already exists on Host Zone (%s)", msg.Name, msg.HostZone)
			k.Logger(ctx).Error(errMsg)
			return nil, sdkerrors.Wrap(types.ErrValidatorAlreadyExists, errMsg)
		}
	}
	// add the validator
	hostZone.Validators = append(validators, &types.Validator{
		Name:           msg.Name,
		Address:        msg.Address,
		Status:         types.Validator_VALIDATOR_STATUS_ACTIVE,
		CommissionRate: msg.Commission,
		DelegationAmt:  0,
		Weight:         msg.Weight,
	})
	k.SetHostZone(ctx, hostZone)
	return &types.MsgAddValidatorResponse{}, nil
}
