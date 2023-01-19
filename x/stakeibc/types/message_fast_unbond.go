package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgFastUnbond = "fast_unbond"

var _ sdk.Msg = &MsgFastUnbond{}

func NewMsgFastUnbond(creator string, amount sdkmath.Int, hostZone string) *MsgFastUnbond {
	return &MsgFastUnbond{
		Creator:  creator,
		Amount:   amount,
		HostZone: hostZone,
	}
}

func (msg *MsgFastUnbond) Route() string {
	return RouterKey
}

func (msg *MsgFastUnbond) Type() string {
	return TypeMsgFastUnbond
}

func (msg *MsgFastUnbond) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgFastUnbond) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgFastUnbond) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return sdkerrors.Wrapf(ErrInvalidAmount, "invalid amount (%v) must be positive and nonzero", msg.Amount)
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return sdkerrors.Wrapf(ErrRequiredFieldEmpty, "host zone cannot be empty")
	}
	return nil
}
