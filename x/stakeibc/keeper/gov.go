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

func (k Keeper) DeleteValidatorsProposal(ctx sdk.Context, msg *types.DeleteValidatorsProposal) error {
	for _, valAddr := range msg.ValAddrs {
		if err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, valAddr); err != nil {
			return err
		}
	}

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
