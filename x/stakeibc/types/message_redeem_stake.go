package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRedeemStake = "redeem_stake"

var _ sdk.Msg = &MsgRedeemStake{}

func NewMsgRedeemStake(creator string, amount sdkmath.Int, hostZone string, receiver string) *MsgRedeemStake {
	return &MsgRedeemStake{
		Creator:  creator,
		Amount:   amount,
		HostZone: hostZone,
		Receiver: receiver,
	}
}

func (msg *MsgRedeemStake) Route() string {
	return RouterKey
}

func (msg *MsgRedeemStake) Type() string {
	return TypeMsgRedeemStake
}

func (msg *MsgRedeemStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRedeemStake) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate host zone is not empty
	// we check validity in the RedeemState function
	if msg.Receiver == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "receiver cannot be empty")
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount (%v)", msg.Amount)
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "host zone cannot be empty")
	}
	return nil
}
