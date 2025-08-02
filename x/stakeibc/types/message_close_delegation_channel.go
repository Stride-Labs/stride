package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v27/utils"
)

const TypeMsgCloseDelegationChannel = "close_delegation_channel"

var _ sdk.Msg = &MsgCloseDelegationChannel{}

func NewMsgCloseDelegationChannel(creator, chainId string) *MsgCloseDelegationChannel {
	return &MsgCloseDelegationChannel{
		Creator: creator,
		ChainId: chainId,
	}
}

func (msg *MsgCloseDelegationChannel) Route() string {
	return RouterKey
}

func (msg *MsgCloseDelegationChannel) Type() string {
	return TypeMsgCloseDelegationChannel
}

func (msg *MsgCloseDelegationChannel) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCloseDelegationChannel) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	if msg.ChainId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chain ID must be specified")
	}

	return nil
}
