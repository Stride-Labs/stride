package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgClaimDaily{}, "airdrop/MsgClaimDaily")
	legacy.RegisterAminoMsg(cdc, &MsgClaimEarly{}, "airdrop/MsgClaimEarly")
	legacy.RegisterAminoMsg(cdc, &MsgClaimAndStake{}, "airdrop/MsgClaimAndStake")
	legacy.RegisterAminoMsg(cdc, &MsgCreateAirdrop{}, "airdrop/MsgCreateAirdrop")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateAirdrop{}, "airdrop/MsgUpdateAirdrop")
	legacy.RegisterAminoMsg(cdc, &MsgAddAllocations{}, "airdrop/MsgAddAllocations")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateUserAllocation{}, "airdrop/MsgUpdateUserAllocation")
	legacy.RegisterAminoMsg(cdc, &MsgLinkAddresses{}, "airdrop/MsgLinkAddresses")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgClaimDaily{},
		&MsgClaimEarly{},
		&MsgClaimAndStake{},
		&MsgCreateAirdrop{},
		&MsgUpdateAirdrop{},
		&MsgAddAllocations{},
		&MsgUpdateUserAllocation{},
		&MsgLinkAddresses{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	cryptocodec.RegisterCrypto(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)
}
