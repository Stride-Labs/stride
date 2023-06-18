package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AddValidators(goCtx context.Context, msg *types.MsgAddValidators) (*types.MsgAddValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, false); err != nil {
			return nil, err
		}
	}

	return &types.MsgAddValidatorsResponse{}, nil
}
