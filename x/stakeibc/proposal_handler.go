package stakeibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func NewStakeibcProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorProposal:
			return handleAddValidatorProposal(ctx, k, c)

		default:
			return fmt.Errorf("unrecognized stakeibc proposal content type: %T: unknown request", c)
		}
	}
}

func handleAddValidatorProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.AddValidatorProposal) error {
	return k.AddValidatorProposal(ctx, proposal)
}
