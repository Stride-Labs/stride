package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgClaimDaily{}, "airdrop/MsgClaimDaily")
	legacy.RegisterAminoMsg(cdc, &MsgClaimEarly{}, "airdrop/MsgClaimEarly")
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
		&MsgCreateAirdrop{},
		&MsgUpdateAirdrop{},
		&MsgAddAllocations{},
		&MsgUpdateUserAllocation{},
		&MsgLinkAddresses{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
