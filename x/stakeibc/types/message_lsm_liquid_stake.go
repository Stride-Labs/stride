package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgLSMLiquidStake = "lsm_liquid_stake"

var _ sdk.Msg = &MsgLSMLiquidStake{}

func NewMsgLSMLiquidStake(creator string, amount sdkmath.Int, lsmTokenIbcDenom string) *MsgLSMLiquidStake {
	return &MsgLSMLiquidStake{
		Creator:          creator,
		Amount:           amount,
		LsmTokenIbcDenom: lsmTokenIbcDenom,
	}
}

func (msg *MsgLSMLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLSMLiquidStake) Type() string {
	return TypeMsgLSMLiquidStake
}

func (msg *MsgLSMLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLSMLiquidStake) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount (%v)", msg.Amount)
	}
	// validate host denom is not empty
	if msg.LsmTokenIbcDenom == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "LSM token denom cannot be empty")
	}
	// lsm token denom must be a valid asset denom matching regex
	if err := sdk.ValidateDenom(msg.LsmTokenIbcDenom); err != nil {
		return errorsmod.Wrapf(err, "invalid LSM token denom")
	}
	return nil
}
