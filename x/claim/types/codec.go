package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSetAirdropAllocations{}, "claim/MsgSetAirdropAllocations")
	legacy.RegisterAminoMsg(cdc, &MsgClaimFreeAmount{}, "claim/MsgClaimFreeAmount")
	legacy.RegisterAminoMsg(cdc, &MsgCreateAirdrop{}, "claim/MsgCreateAirdrop")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteAirdrop{}, "claim/MsgDeleteAirdrop")
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
