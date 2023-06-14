package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) AddValidatorsProposal(ctx sdk.Context, msg *types.AddValidatorsProposal) error {
	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) DeleteValidatorsProposal(ctx sdk.Context, msg *types.DeleteValidatorsProposal) error {
	for _, valAddr := range msg.ValAddrs {
		if err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, valAddr); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) ChangeValidatorWeightsProposal(ctx sdk.Context, msg *types.ChangeValidatorWeightsProposal) error {
	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone %s not found", msg.HostZone))
		return types.ErrInvalidHostZone
	}

	weights := make(map[string]uint64)
	for i, valAddr := range msg.ValAddrs {
		weights[valAddr] = msg.Weights[i]
	}

	validators := hostZone.Validators
	for _, validator := range validators {
		weight, ok := weights[validator.GetAddress()]
		if ok {
			// when changing a weight from 0 to non-zero, make sure we have space in the val set for this new validator
			if validator.Weight == 0 && weight > 0 {
				err := k.ConfirmValSetHasSpace(ctx, validators)
				if err != nil {
					return errorsmod.Wrap(types.ErrMaxNumValidators, "cannot set val weight from zero to nonzero on host zone")
				}
			}
			validator.Weight = weight
		}
	}
	k.SetHostZone(ctx, hostZone)
	return nil
}

func (k Keeper) RegisterHostZoneProposal(ctx sdk.Context, proposal *types.RegisterHostZoneProposal) error {
	return k.RegisterHostZone(
		ctx,
		proposal.ConnectionId,
		proposal.HostDenom,
		proposal.TransferChannelId,
		proposal.Bech32Prefix,
		proposal.IbcDenom,
		proposal.UnbondingFrequency,
		proposal.MinRedemptionRate,
		proposal.MaxRedemptionRate,
	)
}
