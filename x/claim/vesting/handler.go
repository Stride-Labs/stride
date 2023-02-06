package vesting

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	legacysdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/Stride-Labs/stride/v5/x/claim/vesting/types"
)

// NewHandler returns a handler for x/auth message types.
func NewHandler(ak keeper.AccountKeeper, bk types.BankKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		default:
			return nil, sdkerrors.Wrapf(legacysdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
