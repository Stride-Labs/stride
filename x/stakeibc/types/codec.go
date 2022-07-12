package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLiquidStake{}, "stakeibc/LiquidStake", nil)
	cdc.RegisterConcrete(&MsgRegisterAccount{}, "stakeibc/RegisterAccount", nil)
	cdc.RegisterConcrete(&MsgSubmitTx{}, "stakeibc/SubmitTx", nil)
	cdc.RegisterConcrete(&MsgRegisterHostZone{}, "stakeibc/RegisterHostZone", nil)
	cdc.RegisterConcrete(&MsgRedeemStake{}, "stakeibc/RedeemStake", nil)
	cdc.RegisterConcrete(&MsgClaimUndelegatedTokens{}, "stakeibc/ClaimUndelegatedTokens", nil)
	cdc.RegisterConcrete(&MsgRebalanceValidators{}, "stakeibc/RebalanceValidators", nil)
	cdc.RegisterConcrete(&MsgAddValidator{}, "stakeibc/AddValidator", nil)
	cdc.RegisterConcrete(&MsgChangeValidatorWeight{}, "stakeibc/ChangeValidatorWeight", nil)
	cdc.RegisterConcrete(&MsgDeleteValidator{}, "stakeibc/DeleteValidator", nil)
	cdc.RegisterConcrete(&MsgSetNumValidators{}, "stakeibc/SetNumValidators", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgRegisterAccount{},
		&MsgSubmitTx{},
		&MsgRegisterHostZone{},
		&MsgRedeemStake{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil))
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgClaimUndelegatedTokens{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRebalanceValidators{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddValidator{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgChangeValidatorWeight{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteValidator{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetNumValidators{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
