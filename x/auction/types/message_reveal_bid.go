package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRevealBid = "reveal_bid"

var _ sdk.Msg = &MsgSubmitSealedBid{}

func NewMsgRevealBid(creator string, zone string, poolID uint64, bid string, salt string) *MsgRevealBid {
	return &MsgRevealBid{
		Creator: creator,
		Zone:    zone,
		PoolID:  poolID,
		Bid:     bid,
		Salt:    salt,
	}
}

func (msg *MsgRevealBid) Route() string {
	return RouterKey
}

func (msg *MsgRevealBid) Type() string {
	return TypeMsgStartAuction
}

func (msg *MsgRevealBid) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRevealBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRevealBid) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
