package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRestoreInterchainAccount = "restore_interchain_account"

var _ sdk.Msg = &MsgRestoreInterchainAccount{}

func NewMsgRestoreInterchainAccount(creator, chainId, connectionId, owner string) *MsgRestoreInterchainAccount {
	return &MsgRestoreInterchainAccount{
		Creator:      creator,
		ChainId:      chainId,
		ConnectionId: connectionId,
		AccountOwner: owner,
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

func (msg *MsgRestoreInterchainAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.ChainId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chain ID must be specified")
	}
	if !strings.HasPrefix(msg.ConnectionId, "connection-") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "connection ID must be specified")
	}
	if msg.AccountOwner == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ICA account owner must be specified")
	}
	if !strings.HasPrefix(msg.AccountOwner, msg.ChainId) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ICA account owner does not match chain ID")
	}
	return nil
}
