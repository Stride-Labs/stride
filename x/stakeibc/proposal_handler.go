package stakeibc

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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
