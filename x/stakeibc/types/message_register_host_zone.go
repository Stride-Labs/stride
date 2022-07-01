package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRegisterHostZone = "register_host_zone"

var _ sdk.Msg = &MsgRegisterHostZone{}

func NewMsgRegisterHostZone(creator string, connectionId string, hostDenom string, ibcDenom string, transferChannelId string, unbondingFrequency uint64) *MsgRegisterHostZone {
	return &MsgRegisterHostZone{
		Creator:            creator,
		ConnectionId:       connectionId,
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

func (msg *MsgRegisterHostZone) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
