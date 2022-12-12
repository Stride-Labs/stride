package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgAddRateLimit
const TypeMsgAddRateLimit = "add_rate_limit"

var _ sdk.Msg = &MsgAddRateLimit{}

func NewMsgAddRateLimit(creator string, denom string, channelId string, maxPercentSend uint64, maxPercentRecv uint64, durationHours uint64) *MsgAddRateLimit {
	return &MsgAddRateLimit{
		Creator:        creator,
		Denom:          denom,
		ChannelId:      channelId,
		MaxPercentSend: maxPercentSend,
		MaxPercentRecv: maxPercentRecv,
		DurationHours:  durationHours,
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

	if msg.MaxPercentRecv == 0 && msg.MaxPercentSend == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "either the max send or max receive threshold must be greater than 0")
	}

	if msg.DurationHours == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "duration can not be zero")
	}
	return nil
}
