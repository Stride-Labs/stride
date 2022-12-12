package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgAddRateLimit
const TypeMsgAddRateLimit = "add_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgAddRateLimit(creator string, denom string, channelId string, maxPercentSend uint64, maxPercentRecv uint64, durationMinutes uint64) *MsgAddRateLimit {
	return &MsgAddRateLimit{
		Creator:         creator,
		Denom:           denom,
		ChannelId:       channelId,
		MaxPercentSend:  maxPercentSend,
		MaxPercentRecv:  maxPercentRecv,
		DurationMinutes: durationMinutes,
	}
}

func (msg *MsgAddRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgAddRateLimit) Type() string {
	return TypeMsgAddRateLimit
}

func (msg *MsgAddRateLimit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Denom == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid denom")
	}

	if !strings.HasPrefix(msg.ChannelId, "channel-") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid channel-id")
	}

	if msg.MaxPercentRecv > 100 || msg.MaxPercentSend > 100 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "percent must be between 0 and 100 (inclusively)")
	}

	if msg.DurationMinutes == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}
	return nil
}

// Msg type for MsgUpdateRateLimit
const TypeMsgUpdateRateLimit = "update_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgUpdateRateLimit(creator string, pathId string, maxPercentSend uint64, maxPercentRecv uint64, durationMinutes uint64) *MsgUpdateRateLimit {
	return &MsgUpdateRateLimit{
		Creator:         creator,
		PathId:          pathId,
		MaxPercentSend:  maxPercentSend,
		MaxPercentRecv:  maxPercentRecv,
		DurationMinutes: durationMinutes,
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

	if msg.DurationMinutes == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}
	return nil
}

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

// Msg type for MsgResetRateLimit
const TypeMsgResetRateLimit = "reset_rate_limit"

var _ sdk.Msg = &MsgResetRateLimit{}

func NewMsgResetRateLimit(creator string, pathId string) *MsgResetRateLimit {
	return &MsgResetRateLimit{
		Creator: creator,
		PathId:  pathId,
	}
}

func (msg *MsgResetRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgResetRateLimit) Type() string {
	return TypeMsgResetRateLimit
}

func (msg *MsgResetRateLimit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgResetRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResetRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.PathId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pathId")
	}
	return nil
}
