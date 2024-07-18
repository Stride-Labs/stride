package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/Stride-Labs/stride/v22/utils"
)

const TypeMsgCloseICAChannel = "close_ica_channel"

var (
	_ sdk.Msg            = &MsgCloseICAChannel{}
	_ legacytx.LegacyMsg = &MsgCloseICAChannel{}
)

func NewMsgCloseICAChannel(creator, channelId, portId string) *MsgCloseICAChannel {
	return &MsgCloseICAChannel{
		Creator:   creator,
		ChannelId: channelId,
		PortId:    portId,
	}
}

func (msg *MsgCloseICAChannel) Route() string {
	return RouterKey
}

func (msg *MsgCloseICAChannel) Type() string {
	return TypeMsgCloseICAChannel
}

func (msg *MsgCloseICAChannel) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCloseICAChannel) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCloseICAChannel) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	if msg.ChannelId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "channel ID must be specified")
	}
	if msg.PortId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "port ID must be specified")
	}
	if err := ValidateChannelId(msg.ChannelId); err != nil {
		return err
	}

	return nil
}
