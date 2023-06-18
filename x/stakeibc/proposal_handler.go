package stakeibc

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v10/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func NewStakeibcProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddValidatorsProposal:
			return k.AddValidatorsProposal(ctx, c)

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized stakeibc proposal content type: %T", c)
		}
	}
}
