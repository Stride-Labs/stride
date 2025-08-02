package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgAddOracle{}, "icaoracle/MsgAddOracle")
	legacy.RegisterAminoMsg(cdc, &MsgInstantiateOracle{}, "icaoracle/MsgInstantiateOracle")
	legacy.RegisterAminoMsg(cdc, &MsgRestoreOracleICA{}, "icaoracle/MsgRestoreOracleICA")
	legacy.RegisterAminoMsg(cdc, &MsgToggleOracle{}, "icaoracle/MsgToggleOracle")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveOracle{}, "icaoracle/MsgRemoveOracle")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddOracle{},
		&MsgInstantiateOracle{},
		&MsgRestoreOracleICA{},
		&MsgToggleOracle{},
		&MsgRemoveOracle{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
