package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgLSMLiquidStake = "lsm_liquid_stake"

var _ sdk.Msg = &MsgLSMLiquidStake{}

func NewMsgLSMLiquidStake(creator string, amount sdkmath.Int, lsmTokenDenom string) *MsgLSMLiquidStake {
	return &MsgLSMLiquidStake{
		Creator:       creator,
		Amount:        amount,
		LsmTokenDenom: lsmTokenDenom,
	}
}

func (msg *MsgLSMLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLSMLiquidStake) Type() string {
	return TypeMsgLSMLiquidStake
}

func (msg *MsgLSMLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLSMLiquidStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLSMLiquidStake) ValidateBasic() error {
	// TODO [LSM]
	return nil
}
