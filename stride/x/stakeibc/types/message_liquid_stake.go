package types

import (
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
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
