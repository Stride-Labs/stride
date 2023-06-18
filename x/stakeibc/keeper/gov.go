package keeper

import (
	"github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddValidatorsProposal(ctx sdk.Context, msg *types.AddValidatorsProposal) error {
	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, true); err != nil {
			return err
		}
	}

	return nil
}
