package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k Keeper) AddValidatorProposal(ctx sdk.Context, msg *types.AddValidatorProposal) error {
	fmt.Println("ABOUT TO ADD HOST ZONE")

	hz, found := k.GetHostZone(ctx, "GAIA")
	if !found {
		return errors.New("not found")
	}

	hz.HostDenom = "fake-" + msg.HostZone

	k.SetHostZone(ctx, hz)

	fmt.Println("DONE ADDING HOST ZONE")
	return nil
}

func (k Keeper) DeleteValidatorProposal(ctx sdk.Context, msg *types.DeleteValidatorProposal) error {
	fmt.Println("ABOUT TO REMOVE HOST ZONE")

	k.RemoveHostZone(ctx, "GAIA")

	fmt.Println("DONE ADDING HOST ZONE")
	return nil
}
