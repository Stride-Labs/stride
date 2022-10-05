package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDepositAirdrop = "deposit_airdrop"

var _ sdk.Msg = &MsgDepositAirdrop{}

func NewMsgDepositAirdrop(distributor string, airdropAmount sdk.Coins) *MsgDepositAirdrop {
	return &MsgDepositAirdrop{
		Distributor:   distributor,
		AirdropAmount: airdropAmount,
	}
}

func (msg *MsgDepositAirdrop) Route() string {
	return RouterKey
}

func (msg *MsgDepositAirdrop) Type() string {
	return TypeMsgDepositAirdrop
}

func (msg *MsgDepositAirdrop) GetSigners() []sdk.AccAddress {
	distributor, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{distributor}
}

func (msg *MsgDepositAirdrop) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositAirdrop) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address (%s)", err)
	}

	if msg.AirdropAmount.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "empty coin (%s)", err)
	}

	return nil
}
