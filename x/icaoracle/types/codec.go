package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govcodec "github.com/cosmos/cosmos-sdk/x/gov/codec"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgAddOracle{}, "icaoracle/AddOracle")
	legacy.RegisterAminoMsg(cdc, &MsgInstantiateOracle{}, "icaoracle/InstantiateOracle")
	legacy.RegisterAminoMsg(cdc, &MsgRestoreOracleICA{}, "icaoracle/RestoreOracleICA")
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

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz  and gov Amino codec so that this can later be
	// used to properly serialize MsgSubmitProposal instances
	RegisterLegacyAminoCodec(govcodec.Amino)
}
