package stakeibc

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/keeper"

	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

func NewStakeibcProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorProposal:
			return handleAddValidatorProposal(ctx, k, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized stakeibc proposal content type: %T", c)
		}
	}
}

func handleAddValidatorProposal(ctx sdk.Context, k keeper.Keeper, proposal *types.AddValidatorProposal) error {
	return k.AddValidatorProposal(ctx, proposal)
}
