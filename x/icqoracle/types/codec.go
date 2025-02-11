package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govcodec "github.com/cosmos/cosmos-sdk/x/gov/codec"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types/gamm"
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

	registry.RegisterInterface(
		"osmosis.gamm.v1beta1.PoolI",
		(*gamm.PoolI)(nil),
		&gamm.OsmosisGammPool{},
	)

	registry.RegisterInterface(
		"osmosis.gamm.v1beta1.CFMMPoolI",
		(*gamm.CFMMPoolI)(nil),
		&gamm.OsmosisGammPool{},
	)

	proto.RegisterType((*gamm.OsmosisGammPool)(nil), "osmosis.gamm.v1beta1.Pool")

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

	// Register all Amino interfaces and concrete types on the authz and gov Amino codec so that this can later be
	// used to properly serialize MsgSubmitProposal instances
	RegisterLegacyAminoCodec(govcodec.Amino)
}
