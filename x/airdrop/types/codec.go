package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClaim{}, "airdrop/MsgClaim", nil)
	cdc.RegisterConcrete(&MsgClaimEarly{}, "airdrop/MsgClaimEarly", nil)
	cdc.RegisterConcrete(&MsgClaimAndStake{}, "airdrop/MsgClaimAndStake", nil)
	// TODO[airdrop]: add admin messages
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgClaim{},
		&MsgClaimEarly{},
		&MsgClaimAndStake{},
		// TODO[airdrop]: add admin messages
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
