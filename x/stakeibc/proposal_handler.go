package stakeibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// NewAddValidatorHandler creates a new governance Handler for a AddValidatorProposal
func NewStakeibcProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorProposal:
			_, err := k.AddValidatorProposal(ctx, c)
			if err != nil {
				return err
			}
			return nil

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized addValidator proposal content type: %T", c)
		}
	}
}
