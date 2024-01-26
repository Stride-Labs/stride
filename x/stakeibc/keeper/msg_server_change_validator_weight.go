package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeights) (*types.MsgChangeValidatorWeightsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	for _, weightChange := range msg.ValidatorWeights {

		validatorFound := false
		for _, validator := range hostZone.Validators {
			if validator.Address == weightChange.Address {
				validator.Weight = weightChange.Weight
				k.SetHostZone(ctx, hostZone)

				validatorFound = true
				break
			}
		}

		if !validatorFound {
			return nil, types.ErrValidatorNotFound
		}
	}

	// Confirm the new weights wouldn't cause any validator to exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, hostZone.Validators); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to change validator weight")
	}

	return &types.MsgChangeValidatorWeightsResponse{}, nil
}
