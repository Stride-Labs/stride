package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgUpdateRateLimit
const TypeMsgUpdateRateLimit = "update_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgUpdateRateLimit(creator string, pathId string, maxPercentSend uint64, maxPercentRecv uint64, durationHours uint64) *MsgUpdateRateLimit {
	return &MsgUpdateRateLimit{
		Creator:        creator,
		PathId:         pathId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
	}
}

func (msg *MsgUpdateRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRateLimit) Type() string {
	return TypeMsgUpdateRateLimit
}

func (msg *MsgUpdateRateLimit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.PathId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pathId")
	}

	if msg.MaxPercentRecv > 100 || msg.MaxPercentSend > 100 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "percent must be between 0 and 100 (inclusively)")
	}

	if msg.DurationHours == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}
	return nil
}
