package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v22/utils"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/migrations/v2/types"
)

const TypeMsgCloseDelegationChannel = "close_delegation_channel"

var (
	_ sdk.Msg            = &MsgCloseDelegationChannel{}
	_ legacytx.LegacyMsg = &MsgCloseDelegationChannel{}
)

func NewMsgCloseDelegationChannel(creator, channelId, portId string) *MsgCloseDelegationChannel {
	return &MsgCloseDelegationChannel{
		Creator:   creator,
		ChannelId: channelId,
		PortId:    portId,
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

func (msg *MsgCloseDelegationChannel) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCloseDelegationChannel) ValidateBasic() error {
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
	if !strings.HasPrefix(msg.PortId, icatypes.ControllerPortPrefix) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "channel must be an ICA channel")
	}
	if !strings.HasSuffix(msg.PortId, types.ICAAccountType_DELEGATION.String()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "channel must be the delegation ICA channel")
	}

	return nil
}
