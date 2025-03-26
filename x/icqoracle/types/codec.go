package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRegisterTokenPriceQuery{}, "icqoracle/MsgRegisterTokenPriceQuery")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveTokenPriceQuery{}, "icqoracle/MsgRemoveTokenPriceQuery")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "icqoracle/MsgUpdateParams")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterTokenPriceQuery{},
		&MsgRemoveTokenPriceQuery{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
