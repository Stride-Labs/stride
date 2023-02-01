package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (msg *MsgRedeemStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemStake) ValidateBasic() error {
	// check valid creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s): %s", err, ErrInvalidAddress.Error())
	}
	// validate host zone is not empty
	// we check validity in the RedeemState function
	if msg.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty: %s", ErrRequiredFieldEmpty.Error())
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("invalid amount (%v): %s", msg.Amount, ErrInvalidRequest.Error())
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return fmt.Errorf("host zone cannot be empty: %s", ErrRequiredFieldEmpty.Error())
	}
	return nil
}
