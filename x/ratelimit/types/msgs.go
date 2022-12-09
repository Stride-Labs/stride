package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgSetQuota
const TypeMsgSetQuota = "set_quota"

var _ sdk.Msg = &MsgSetQuota{}

func NewMsgSetQuota(creator string, name string, maxPercentSend uint64, maxPercentRecv uint64, durationMinutes uint64) *MsgSetQuota {
	return &MsgSetQuota{
		Creator:         creator,
		Name:            name,
		MaxPercentSend:  maxPercentSend,
		MaxPercentRecv:  maxPercentRecv,
		DurationMinutes: durationMinutes,
	}
}

func (msg *MsgSetQuota) Route() string {
	return RouterKey
}

func (msg *MsgSetQuota) Type() string {
	return TypeMsgSetQuota
}

func (msg *MsgSetQuota) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetQuota) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetQuota) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Name == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quota name not set")
	}

	if msg.MaxPercentRecv == 0 && msg.MaxPercentSend == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "only one of recv and send percent can be zero")
	}

	if msg.DurationMinutes == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}

	return nil
}

// Msg type for MsgRemoveQuota
const TypeMsgRemoveQuota = "remove_quota"

var _ sdk.Msg = &MsgRemoveQuota{}

func NewMsgRemoveQuota(creator string, name string) *MsgRemoveQuota {
	return &MsgRemoveQuota{
		Creator: creator,
		Name:    name,
	}
}

func (msg *MsgRemoveQuota) Route() string {
	return RouterKey
}

func (msg *MsgRemoveQuota) Type() string {
	return TypeMsgRemoveQuota
}

func (msg *MsgRemoveQuota) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveQuota) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveQuota) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Name == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "quota name not set")
	}
	return nil
}

// Msg type for MsgAddRateLimit
const TypeMsgAddRateLimit = "add_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgAddRateLimit(creator string, denom string, channel string, quotaName string) *MsgAddRateLimit {
	return &MsgAddRateLimit{
		Creator:   creator,
		Denom:     denom,
		Channel:   channel,
		QuotaName: quotaName,
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
	// TODO:
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
	// TODO:
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
	// TODO:
	return nil
}
