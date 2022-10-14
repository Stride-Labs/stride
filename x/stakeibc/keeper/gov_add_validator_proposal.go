package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k Keeper) AddValidatorProposal(ctx sdk.Context, msg *types.AddValidatorProposal) (*types.MsgAddValidatorResponse, error) {
	fmt.Println("IN PROPOSAL FUNCTION")
	return &types.MsgAddValidatorResponse{}, nil
}
