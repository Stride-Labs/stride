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
)

var (
	_ sdk.Msg       = &MsgBurn{}
	_ sdk.LegacyMsg = &MsgBurn{}
)

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
