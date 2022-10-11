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
	minWeight := ^uint64(0) >> 1
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
		// calc the min weight to assign to new validator
		if validator.Weight < minWeight {
			minWeight = validator.Weight
		}
	}
	// if the validator was added via governance, set it weight by default to the min val weight on the host zone
	var wgt uint64
	if msg.Creator == "GOV" {
		wgt = minWeight
	} else {
		wgt = msg.Weight
	}
	// add the validator
	hostZone.Validators = append(validators, &types.Validator{
		Name:           msg.Name,
		Address:        msg.Address,
		Status:         types.Validator_Active,
		CommissionRate: msg.Commission,
		DelegationAmt:  0,
		Weight:         wgt,
	})
	k.SetHostZone(ctx, hostZone)
	return &types.MsgAddValidatorResponse{}, nil
}
