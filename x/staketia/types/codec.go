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
	legacy.RegisterAminoMsg(cdc, &MsgLiquidStake{}, "stride/MsgLiquidStake")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemStake{}, "stride/MsgRedeemStake")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmDelegation{}, "stride/MsgConfirmDelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUndelegation{}, "stride/MsgConfirmUndelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUnbondedTokenSweep{}, "stride/MsgConfirmUnbondedTokenSweep")
	legacy.RegisterAminoMsg(cdc, &MsgAdjustDelegatedBalance{}, "stride/MsgAdjustDelegatedBalance")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateInnerRedemptionRateBounds{}, "stride/MsgUpdateRedemptionRateBounds")
	legacy.RegisterAminoMsg(cdc, &MsgResumeHostZone{}, "stride/MsgResumeHostZone")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgRedeemStake{},
		&MsgConfirmDelegation{},
		&MsgConfirmUndelegation{},
		&MsgConfirmUnbondedTokenSweep{},
		&MsgAdjustDelegatedBalance{},
		&MsgUpdateInnerRedemptionRateBounds{},
		&MsgResumeHostZone{},
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

	// Register all Amino interfaces and concrete types on the authz and gov Amino codec so that this can later be
	// used to properly serialize MsgSubmitProposal instances
	RegisterLegacyAminoCodec(govcodec.Amino)
}
