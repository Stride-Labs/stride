package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRedeemStake = "redeem_stake"

var _ sdk.Msg = &MsgRedeemStake{}

func NewMsgRedeemStake(creator string, amount uint64, hostZone string, receiver string) *MsgRedeemStake {
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

func (msg *MsgRedeemStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemStake) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate host zone is not empty
	// we check validity in the RedeemState function
	if msg.Receiver == "" {
		return sdkerrors.Wrapf(ErrRequiredFieldEmpty, "receiver cannot be empty")
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount <= 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount (%d)", msg.Amount)
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return sdkerrors.Wrapf(ErrRequiredFieldEmpty, "host zone cannot be empty")
	}
	// math.MaxInt64 == 1<<63 - 1
	if !(msg.Amount < (1<<63 - 1)) {
		return sdkerrors.Wrapf(ErrInvalidAmount, "amount liquid staked must be less than math.MaxInt64 %d", 1<<63-1)
	}
	return nil
}
