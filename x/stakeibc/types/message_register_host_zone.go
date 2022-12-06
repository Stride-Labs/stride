package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgRegisterHostZone = "register_host_zone"

var _ sdk.Msg = &MsgRegisterHostZone{}

func NewMsgRegisterHostZone(creator string, connectionId string, bech32prefix string, hostDenom string, ibcDenom string, transferChannelId string, unbondingFrequency uint64) *MsgRegisterHostZone {
	return &MsgRegisterHostZone{
		Creator:            creator,
		ConnectionId:       connectionId,
		Bech32Prefix:       bech32prefix,
		HostDenom:          hostDenom,
		IbcDenom:           ibcDenom,
		TransferChannelId:  transferChannelId,
		UnbondingFrequency: unbondingFrequency,
	}
}

func (msg *MsgRegisterHostZone) Route() string {
	return RouterKey
}

func (msg *MsgRegisterHostZone) Type() string {
	return TypeMsgRegisterHostZone
}

func (msg *MsgRegisterHostZone) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterHostZone) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// TODO(TEST-112) add validation on bech32prefix upon zone creation
func (msg *MsgRegisterHostZone) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// VALIDATE DENOMS
	// host denom cannot be empty
	if msg.HostDenom == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.HostDenom); err != nil {
		return err
	}

	// ibc denom cannot be empty and must begin with "ibc"
	if msg.IbcDenom == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "ibc denom cannot be empty")
	}
	if !strings.HasPrefix(msg.IbcDenom, "ibc") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "ibc denom must begin with 'ibc'")
	}
	// ibc denom must be valid
	err = ibctransfertypes.ValidateIBCDenom(msg.IbcDenom)
	if err != nil {
		return err
	}
	// bech32 prefix must be non-empty (we validate it fully in msg_server)
	if strings.TrimSpace(msg.Bech32Prefix) == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "bech32 prefix must be non-empty")
	}
	// connection id cannot be empty and must begin with "connection"
	if msg.ConnectionId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "connection id cannot be empty")
	}
	if !strings.HasPrefix(msg.ConnectionId, "connection") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "connection id must begin with 'connection'")
	}
	// transfer channel id cannot be empty
	if msg.TransferChannelId == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "transfer channel id cannot be empty")
	}
	// transfer channel id must begin with "channel"
	if !strings.HasPrefix(msg.TransferChannelId, "channel") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "transfer channel id must begin with 'channel'")
	}
	// unbonding frequency must be positive nonzero
	if msg.UnbondingFrequency < 1 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unbonding frequency must be greater than zero")
	}

	return nil
}

func NewZoneAddress(chainId string) sdk.AccAddress {
	key := append([]byte("zone"), []byte(chainId)...)
	return address.Module(ModuleName, key)
}
