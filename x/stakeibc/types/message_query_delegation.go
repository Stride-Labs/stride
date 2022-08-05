package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgQueryDelegation = "query_delegation"

var _ sdk.Msg = &MsgQueryDelegation{}

func NewMsgQueryDelegation(creator string, hostzone string, valoper string) *MsgQueryDelegation {
	return &MsgQueryDelegation{
		Creator:  creator,
		Hostzone: hostzone,
		Valoper:  valoper,
	}
}

func (msg *MsgQueryDelegation) Route() string {
	return RouterKey
}

func (msg *MsgQueryDelegation) Type() string {
	return TypeMsgQueryDelegation
}

func (msg *MsgQueryDelegation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgQueryDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgQueryDelegation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
