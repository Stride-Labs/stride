package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AddValidator(goCtx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	// TODO restrict this to governance module. add gov module whitelist hooks more broadly
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, host_zone_found := k.GetHostZone(ctx, msg.HostZone)
	if !host_zone_found {
		k.Logger(ctx).Info(fmt.Sprintf("Host Zone not found %s", msg.HostZone))
	}
	validators := hostZone.Validators
	// check that we don't already have this validator
	for _, validator := range validators {
		if validator.Address == msg.Address {
			k.Logger(ctx).Info(fmt.Sprintf("Validator address %s already exists on Host Zone %s", msg.Address, msg.HostZone))
			return nil, types.ErrValidatorAlreadyExists
		}
		if validator.Name == msg.Name {
			k.Logger(ctx).Info(fmt.Sprintf("Validator name %s already exists on Host Zone %s", msg.Name, msg.HostZone))
			return nil, types.ErrValidatorAlreadyExists
		}
	}
	// add the validator
	validators = append(validators, &types.Validator{
		Name:           msg.Name,
		Address:        msg.Address,
		Status:         "active",
		CommissionRate: msg.Commission,
		DelegationAmt:  0,
		Weight:         msg.Weight,
	})

	return &types.MsgAddValidatorResponse{}, nil
}
