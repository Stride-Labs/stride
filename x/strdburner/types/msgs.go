package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgBurn = "burn"
	TypeMsgLink = "link"
)

// -----------------------------------------------
//                     MsgBurn
// -----------------------------------------------

func NewMsgBurn(burner string, amount sdkmath.Int) *MsgBurn {
	return &MsgBurn{
		Burner: burner,
		Amount: amount,
	}
}

func (msg MsgBurn) Type() string {
	return TypeMsgBurn
}

func (msg MsgBurn) Route() string {
	return RouterKey
}

func (msg *MsgBurn) GetSigners() []sdk.AccAddress {
	burner, err := sdk.AccAddressFromBech32(msg.Burner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{burner}
}

func (msg *MsgBurn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Burner)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	// Min threshold of 1 STRD
	minThreshold := int64(1000000)
	if msg.Amount.LT(sdkmath.NewInt(minThreshold)) {
		return fmt.Errorf("amount (%vustrd) is below 1 STRD minimum", msg.Amount)
	}

	return nil
}

// -----------------------------------------------
//                     MsgLink
// -----------------------------------------------

func NewMsgLink(sender string, linkedAddress string) *MsgLink {
	return &MsgLink{
		Sender:        sender,
		LinkedAddress: linkedAddress,
	}
}

func (msg MsgLink) Type() string {
	return TypeMsgLink
}

func (msg MsgLink) Route() string {
	return RouterKey
}

func (msg *MsgLink) GetSigners() []sdk.AccAddress {
	burner, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{burner}
}

func (msg *MsgLink) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	if msg.LinkedAddress == "" {
		return fmt.Errorf("Linked address cannot be empty")
	}

	return nil
}
