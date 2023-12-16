package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeight) (*types.MsgChangeValidatorWeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	validators := hostZone.Validators
	validatorFound := false
	for _, validator := range validators {
		if validator.Address == msg.ValAddr {
			validatorFound = true

			validator.Weight = msg.Weight
			k.SetHostZone(ctx, hostZone)

			break
		}
	}

	if !validatorFound {
		return nil, types.ErrValidatorNotFound
	}

	if err := k.CheckValidatorWeightsBelowCap(ctx, hostZone.Validators); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to change validator weight")
	}

	return &types.MsgChangeValidatorWeightResponse{}, nil
}
