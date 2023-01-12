package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgClearBalance = "clear_balance"

var _ sdk.Msg = &MsgClearBalance{}

func NewMsgClearBalance(creator string, chainId string, amount sdk.Int, channelId string) *MsgClearBalance {
	return &MsgClearBalance{
		Creator: creator,
		ChainId: chainId,
		Amount:  amount,
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
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// basic checks on host denom
	if len(msg.ChainId) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "chainid is required")
	}

	if msg.Amount.LTE(sdk.ZeroInt()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than 0")
	}
	if isValid := channeltypes.IsValidChannelID(msg.Channel); !isValid {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "channel is invalid")
	}
	return nil
}
