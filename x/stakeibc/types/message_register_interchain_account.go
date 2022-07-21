package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRegisterInterchainAccount = "register_interchain_account"

var _ sdk.Msg = &MsgRegisterInterchainAccount{}

func NewMsgRegisterInterchainAccount(creator string, chainId string, accountType string) *MsgRegisterInterchainAccount {
	return &MsgRegisterInterchainAccount{
		Creator:     creator,
		ChainId:     chainId,
		AccountType: accountType,
	}
}

func (msg *MsgRegisterInterchainAccount) Route() string {
	return RouterKey
}

func (msg *MsgRegisterInterchainAccount) Type() string {
	return TypeMsgRegisterInterchainAccount
}

func (msg *MsgRegisterInterchainAccount) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterInterchainAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterInterchainAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
