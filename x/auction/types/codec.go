package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPlaceBid{}, "auction/MsgPlaceBid")
	legacy.RegisterAminoMsg(cdc, &MsgCreateAuction{}, "auction/MsgCreateAuction")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateAuction{}, "auction/MsgUpdateAuction")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPlaceBid{},
		&MsgCreateAuction{},
		&MsgUpdateAuction{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
