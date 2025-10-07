package types

import (
	"errors"
	"fmt"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgBurn = "burn"
	TypeMsgLink = "link"
)

var (
	_ sdk.Msg = &MsgBurn{}
	_ sdk.Msg = &MsgLink{}

	_ sdk.LegacyMsg = &MsgBurn{}
	_ sdk.LegacyMsg = &MsgLink{}
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
	minThreshold := int64(1_000_000)
	if msg.Amount.LT(sdkmath.NewInt(minThreshold)) {
		return fmt.Errorf("amount (%vustrd) is below 1 STRD minimum", msg.Amount)
	}

	return nil
}

// -----------------------------------------------
//                     MsgLink
// -----------------------------------------------

func NewMsgLink(strideAddress string, linkedAddress string) *MsgLink {
	return &MsgLink{
		StrideAddress: strideAddress,
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
	burner, err := sdk.AccAddressFromBech32(msg.StrideAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{burner}
}

func (msg *MsgLink) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.StrideAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	if msg.LinkedAddress == "" {
		return errors.New("linked address cannot be empty")
	}

	if len(msg.LinkedAddress) > 200 {
		return fmt.Errorf("address must be less than 200 characters, %d provided", len(msg.LinkedAddress))
	}

	// Check if LinkedAddress is alphanumeric
	alphanumericPattern := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !alphanumericPattern.MatchString(msg.LinkedAddress) {
		return fmt.Errorf("linked address must be alphanumeric")
	}

	return nil
}
