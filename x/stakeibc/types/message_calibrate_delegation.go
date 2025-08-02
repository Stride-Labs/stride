package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCalibrateDelegation = "calibrate_delegation"

var _ sdk.Msg = &MsgCalibrateDelegation{}

func NewMsgCalibrateDelegation(creator string, chainid string, valoper string) *MsgCalibrateDelegation {
	return &MsgCalibrateDelegation{
		Creator: creator,
		ChainId: chainid,
		Valoper: valoper,
	}
}

func (msg *MsgCalibrateDelegation) Route() string {
	return RouterKey
}

func (msg *MsgCalibrateDelegation) Type() string {
	return TypeMsgCalibrateDelegation
}

func (msg *MsgCalibrateDelegation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCalibrateDelegation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.ChainId) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chainid is required")
	}
	if len(msg.Valoper) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "valoper is required")
	}
	if !strings.Contains(msg.Valoper, "valoper") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must contrain 'valoper'")
	}

	return nil
}
