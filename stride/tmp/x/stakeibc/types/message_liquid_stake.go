package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgLiquidStake = "liquid_stake"

var _ sdk.Msg = &MsgLiquidStake{}

func NewMsgLiquidStake(creator string, amount int32, denom string) *MsgLiquidStake {
	return &MsgLiquidStake{
		Creator: creator,
		Amount:  amount,
		Denom:   denom,
	}
}

// isIBCToken checks if the token came from the IBC module
// Each IBC token starts with an ibc/ denom, the check is rather simple
func IsIBCToken(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

func (msg *MsgLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLiquidStake) Type() string {
	return TypeMsgLiquidStake
}

func (msg *MsgLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLiquidStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLiquidStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	isIbcToken := IsIBCToken(msg.Denom)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	} else if !isIbcToken {
		return sdkerrors.Wrapf(ErrInvalidToken, "invalid token denom (%s)", msg.Denom)
	}
	return nil
}
