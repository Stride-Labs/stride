package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitOpenBid = "submit_open_bid"

var _ sdk.Msg = &MsgSubmitOpenBid{}

func NewMsgSubmitOpenBid(creator string, zone string, poolID uint64, Bid string) *MsgSubmitOpenBid {
	return &MsgSubmitOpenBid{
		Creator: creator,
		Zone:    zone,
		PoolID:  poolID,
		Bid:     Bid,
	}
}

func (msg *MsgSubmitOpenBid) Route() string {
	return RouterKey
}

func (msg *MsgSubmitOpenBid) Type() string {
	return TypeMsgSubmitOpenBid
}

func (msg *MsgSubmitOpenBid) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitOpenBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitOpenBid) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
