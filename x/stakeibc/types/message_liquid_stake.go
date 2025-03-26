package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgLiquidStake = "liquid_stake"

var _ sdk.Msg = &MsgLiquidStake{}

func NewMsgLiquidStake(creator string, amount sdkmath.Int, hostDenom string) *MsgLiquidStake {
	return &MsgLiquidStake{
		Creator:   creator,
		Amount:    amount,
		HostDenom: hostDenom,
	}
}

func (msg *MsgLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLiquidStake) Type() string {
	return TypeMsgLiquidStake
}

func (msg *MsgLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLiquidStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate amount is positive nonzero
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmount, "amount liquid staked must be positive and nonzero")
	}
	// validate host denom is not empty
	if msg.HostDenom == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "host denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.HostDenom); err != nil {
		return err
	}
	return nil
}
