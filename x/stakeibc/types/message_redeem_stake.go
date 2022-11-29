package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		return fmt.Errorf("invalid creator address (%s): invalid address", err.Error())
	}
	// validate host zone is not empty
	// we check validity in the RedeemState function
	if msg.Receiver == "" {
		return fmt.Errorf("receiver cannot be empty:  %s", ErrRequiredFieldEmpty)
	}
	// ensure amount is a nonzero positive integer
	if msg.Amount <= 0 {
		return fmt.Errorf("invalid amount (%d): invalid request", msg.Amount)
	}
	// validate host zone is not empty
	if msg.HostZone == "" {
		return fmt.Errorf("host zone cannot be empty: %s", ErrRequiredFieldEmpty)
	}
	// math.MaxInt64 == 1<<63 - 1
	if !(msg.Amount < (1<<63 - 1)) {
		return fmt.Errorf("amount liquid staked must be less than math.MaxInt64 %d: %s", 1<<63-1, ErrInvalidAmount)
	}
	return nil
}
