package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k msgServer) AddValidators(goCtx context.Context, msg *types.MsgAddValidators) (*types.MsgAddValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, false); err != nil {
			return nil, err
		}

		// Query and store the validator's exchange rate
		callbackData := []byte{}
		aggressiveTimeout := false
		if err := k.QueryValidatorExchangeRate(ctx, msg.HostZone, validator.Address, callbackData, aggressiveTimeout); err != nil {
			return nil, err
		}
	}

	return &types.MsgAddValidatorsResponse{}, nil
}
