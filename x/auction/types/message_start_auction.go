package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgStartAuction = "start_auction"

var _ sdk.Msg = &MsgStartAuction{}

func NewMsgStartAuction(creator string, zone string, poolID uint64) *MsgStartAuction {
	return &MsgStartAuction{
		Creator: creator,
		Zone:    zone,
		PoolID:  poolID,
	}
}

func (msg *MsgStartAuction) Route() string {
	return RouterKey
}

func (msg *MsgStartAuction) Type() string {
	return TypeMsgStartAuction
}

func (msg *MsgStartAuction) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgStartAuction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgStartAuction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
