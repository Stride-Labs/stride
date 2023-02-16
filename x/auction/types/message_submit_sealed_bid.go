package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSubmitSealedBid = "submit_sealed_bid"

var _ sdk.Msg = &MsgSubmitSealedBid{}

func NewMsgSubmitSealedBid(creator string, zone string, poolID uint64, hashedBid string) *MsgSubmitSealedBid {
	return &MsgSubmitSealedBid{
		Creator:   creator,
		Zone:      zone,
		PoolID:    poolID,
		HashedBid: hashedBid,
	}
}

func (msg *MsgSubmitSealedBid) Route() string {
	return RouterKey
}

func (msg *MsgSubmitSealedBid) Type() string {
	return TypeMsgStartAuction
}

func (msg *MsgSubmitSealedBid) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitSealedBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitSealedBid) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
