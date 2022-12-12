package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty pathId")
	}

	if !strings.Contains(msg.PathId, "/") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid pathId")
	}
	return nil
}
