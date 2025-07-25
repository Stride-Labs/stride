package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func (k Keeper) AddValidatorsProposal(ctx sdk.Context, msg *types.AddValidatorsProposal) error {
	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, true); err != nil {
			return err
		}
	}

	// Confirm none of the validator's exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, msg.HostZone); err != nil {
		return err
	}

	return nil
}

func (k Keeper) ToggleLSMProposal(ctx sdk.Context, msg *types.ToggleLSMProposal) error {
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return types.ErrHostZoneNotFound.Wrap(msg.HostZone)
	}

	hostZone.LsmLiquidStakeEnabled = msg.Enabled
	k.SetHostZone(ctx, hostZone)

	return nil
}
