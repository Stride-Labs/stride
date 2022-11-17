package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSetAirdropAllocations{}, "claim/SetAirdropAllocation", nil)
	cdc.RegisterConcrete(&MsgClaimFreeAmount{}, "claim/ClaimFreeAmount", nil)
	cdc.RegisterConcrete(&MsgCreateAirdrop{}, "claim/MsgCreateAirdrop", nil)
	cdc.RegisterConcrete(&MsgDeleteAirdrop{}, "claim/MsgDeleteAirdrop", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetAirdropAllocations{},
		&MsgClaimFreeAmount{},
		&MsgCreateAirdrop{},
		&MsgDeleteAirdrop{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
	cryptocodec.RegisterCrypto(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)
}
