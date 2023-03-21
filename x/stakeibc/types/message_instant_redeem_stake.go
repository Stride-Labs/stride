package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgInstantRedeemStake = "fast_unbond"

var _ sdk.Msg = &MsgInstantRedeemStake{}

func NewMsgInstantRedeemStake(creator string, amount sdkmath.Int, hostZone string) *MsgInstantRedeemStake {
	return &MsgInstantRedeemStake{
		Creator:  creator,
		Amount:   amount,
		HostZone: hostZone,
	}
}

func (msg *MsgInstantRedeemStake) Route() string {
	return RouterKey
}

func (msg *MsgInstantRedeemStake) Type() string {
	return TypeMsgInstantRedeemStake
}

func (msg *MsgInstantRedeemStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgInstantRedeemStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgInstantRedeemStake) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return sdkerrors.Wrapf(ErrInvalidAmount, "invalid amount (%v) must be positive and nonzero", msg.Amount)
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return sdkerrors.Wrapf(ErrRequiredFieldEmpty, "host zone cannot be empty")
	}
	return nil
}
