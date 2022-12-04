package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k Keeper) AddValidatorProposal(ctx sdk.Context, msg *types.AddValidatorProposal) error {
	addValMsg := &types.MsgAddValidator{
		HostZone:   msg.HostZone,
		Name:       msg.ValidatorName,
		Address:    msg.ValidatorAddress,
		Commission: 0, // TODO: Remove commission field from validator
	}
	return k.AddValidatorToHostZone(ctx, addValMsg, true)
}
