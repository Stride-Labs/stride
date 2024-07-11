package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) Claim(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}

func (k Keeper) ClaimAndStake(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}

func (k Keeper) ClaimEarly(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}
