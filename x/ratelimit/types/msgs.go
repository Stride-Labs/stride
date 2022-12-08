package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Msg type for MsgAddQuota
const TypeMsgAddQuota = "add_quota"

var _ sdk.Msg = &MsgAddQuota{}

func NewMsgAddQuota(name string, maxPercentSend uint64, maxPercentRecv uint64, durationMinutes uint64) *MsgAddQuota {
	return &MsgAddQuota{
		Name:            name,
		MaxPercentSend:  maxPercentSend,
		MaxPercentRecv:  maxPercentRecv,
		DurationMinutes: durationMinutes,
	}
}

func (msg *MsgAddQuota) Route() string {
	return RouterKey
}

func (msg *MsgAddQuota) Type() string {
	return TypeMsgAddQuota
}

func (msg *MsgAddQuota) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
}

func (msg *MsgAddQuota) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddQuota) ValidateBasic() error {
	// TODO:
	return nil
}

// Msg type for MsgRemoveQuota
const TypeMsgRemoveQuota = "remove_quota"

var _ sdk.Msg = &MsgRemoveQuota{}

func NewMsgRemoveQuota(name string) *MsgRemoveQuota {
	return &MsgRemoveQuota{
		Name: name,
	}
}

func (msg *MsgRemoveQuota) Route() string {
	return RouterKey
}

func (msg *MsgRemoveQuota) Type() string {
	return TypeMsgRemoveQuota
}

func (msg *MsgRemoveQuota) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
}

func (msg *MsgRemoveQuota) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveQuota) ValidateBasic() error {
	// TODO:
	return nil
}

// Msg type for MsgAddRateLimit
const TypeMsgAddRateLimit = "add_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgAddRateLimit(denom string, channel string, quotaName string) *MsgAddRateLimit {
	return &MsgAddRateLimit{
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
	allocator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
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

func NewMsgRemoveRateLimit(pathId string) *MsgRemoveRateLimit {
	return &MsgRemoveRateLimit{
		PathId: pathId,
	}
}

func (msg *MsgRemoveRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgRemoveRateLimit) Type() string {
	return TypeMsgRemoveRateLimit
}

func (msg *MsgRemoveRateLimit) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
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

func NewMsgResetRateLimit(pathId string) *MsgResetRateLimit {
	return &MsgResetRateLimit{
		PathId: pathId,
	}
}

func (msg *MsgResetRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgResetRateLimit) Type() string {
	return TypeMsgResetRateLimit
}

func (msg *MsgResetRateLimit) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
}

func (msg *MsgResetRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResetRateLimit) ValidateBasic() error {
	// TODO:
	return nil
}
