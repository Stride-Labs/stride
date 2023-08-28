package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeight) (*types.MsgChangeValidatorWeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	validators := hostZone.Validators
	for _, validator := range validators {
		if validator.Address == msg.ValAddr {
			validator.Weight = msg.Weight
			k.SetHostZone(ctx, hostZone)
			return &types.MsgChangeValidatorWeightResponse{}, nil
		}
	}

	return nil, types.ErrValidatorNotFound
}
