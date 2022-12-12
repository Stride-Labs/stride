package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgRemoveRateLimit
const TypeMsgRemoveRateLimit = "remove_rate_limit"

var _ sdk.Msg = &MsgRemoveRateLimit{}

func NewMsgRemoveRateLimit(creator string, pathId string) *MsgRemoveRateLimit {
	return &MsgRemoveRateLimit{
		Creator: creator,
		PathId:  pathId,
	}
}

func (msg *MsgRemoveRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgRemoveRateLimit) Type() string {
	return TypeMsgRemoveRateLimit
}

func (msg *MsgRemoveRateLimit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.PathId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pathId")
	}
	return nil
}
