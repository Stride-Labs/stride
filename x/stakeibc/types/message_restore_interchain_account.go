package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRestoreInterchainAccount = "register_interchain_account"

var _ sdk.Msg = &MsgRestoreInterchainAccount{}

func NewMsgRestoreInterchainAccount(creator string, chainId string, accountType ICAAccountType) *MsgRestoreInterchainAccount {
	return &MsgRestoreInterchainAccount{
		Creator:     creator,
		ChainId:     chainId,
		AccountType: accountType,
	}
}

func (msg *MsgRestoreInterchainAccount) Route() string {
	return RouterKey
}

func (msg *MsgRestoreInterchainAccount) Type() string {
	return TypeMsgRestoreInterchainAccount
}

func (msg *MsgRestoreInterchainAccount) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRestoreInterchainAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRestoreInterchainAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
