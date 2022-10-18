package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
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

func (k Keeper) DeleteValidatorProposal(ctx sdk.Context, msg *types.DeleteValidatorProposal) error {
	fmt.Println("ABOUT TO REMOVE HOST ZONE")

	k.RemoveHostZone(ctx, "GAIA")

	fmt.Println("DONE ADDING HOST ZONE")
	return nil
}
