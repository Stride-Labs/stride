package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) AddValidatorsProposal(ctx sdk.Context, msg *types.AddValidatorsProposal) error {
	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, true); err != nil {
			return err
		}
	}

	return nil
}
