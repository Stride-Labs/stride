package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgClearBalance = "clear_balance"

var _ sdk.Msg = &MsgClearBalance{}

func NewMsgClearBalance(creator string, chainId string, amount uint64, channelId string) *MsgClearBalance {
	return &MsgClearBalance{
		Creator:   creator,
		ChainId:    chainId,
		Amount: amount,
		Channel: channelId,
	}
}

func (msg *MsgClearBalance) Route() string {
	return RouterKey
}

func (msg *MsgClearBalance) Type() string {
	return TypeMsgClearBalance
}

func (msg *MsgClearBalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgClearBalance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClearBalance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// TODO: add validation
	// Should we let anyone call this?
	// if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
	// 	return err
	// }
	return nil
}
