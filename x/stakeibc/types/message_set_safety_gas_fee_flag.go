package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSetSafetyGasFeeFlag = "set_safety_gas_fee_flag"

var _ sdk.Msg = &MsgSetSafetyGasFeeFlag{}

func NewMsgSetSafetyGasFeeFlag(creator string, isEnabled bool) *MsgSetSafetyGasFeeFlag {
	return &MsgSetSafetyGasFeeFlag{
		Creator:   creator,
		IsEnabled: isEnabled,
	}
}

func (msg *MsgSetSafetyGasFeeFlag) Route() string {
	return RouterKey
}

func (msg *MsgSetSafetyGasFeeFlag) Type() string {
	return TypeMsgSetSafetyGasFeeFlag
}

func (msg *MsgSetSafetyGasFeeFlag) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetSafetyGasFeeFlag) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetSafetyGasFeeFlag) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
