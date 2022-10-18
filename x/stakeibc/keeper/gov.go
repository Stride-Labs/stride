package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k Keeper) AddValidatorProposal(ctx sdk.Context, msg *types.AddValidatorProposal) error {
	fmt.Println("ABOUT TO ADD HOST ZONE")

	k.SetHostZone(ctx, types.HostZone{
		ChainId:   "FAKE",
		HostDenom: "fake",
	})

	fmt.Println("DONE ADDING HOST ZONE")
	return nil
}
