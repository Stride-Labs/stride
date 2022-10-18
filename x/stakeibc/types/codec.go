package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLiquidStake{}, "stakeibc/LiquidStake", nil)
	cdc.RegisterConcrete(&MsgRegisterHostZone{}, "stakeibc/RegisterHostZone", nil)
	cdc.RegisterConcrete(&MsgRedeemStake{}, "stakeibc/RedeemStake", nil)
	cdc.RegisterConcrete(&MsgClaimUndelegatedTokens{}, "stakeibc/ClaimUndelegatedTokens", nil)
	cdc.RegisterConcrete(&MsgRebalanceValidators{}, "stakeibc/RebalanceValidators", nil)
	cdc.RegisterConcrete(&MsgAddValidator{}, "stakeibc/AddValidator", nil)
	cdc.RegisterConcrete(&MsgChangeValidatorWeight{}, "stakeibc/ChangeValidatorWeight", nil)
	cdc.RegisterConcrete(&MsgDeleteValidator{}, "stakeibc/DeleteValidator", nil)
	cdc.RegisterConcrete(&AddValidatorProposal{}, "stakeibc/AddValidatorProposal", nil)
	cdc.RegisterConcrete(&MsgRestoreInterchainAccount{}, "stakeibc/RestoreInterchainAccount", nil)
	cdc.RegisterConcrete(&MsgUpdateValidatorSharesExchRate{}, "stakeibc/UpdateValidatorSharesExchRate", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgClearBalance{},
		&MsgRegisterHostZone{},
		&MsgRedeemStake{},
		&MsgClaimUndelegatedTokens{},
		&MsgRebalanceValidators{},
		&MsgAddValidator{},
		&MsgChangeValidatorWeight{},
		&MsgDeleteValidator{},
		&MsgRestoreInterchainAccount{},
		&MsgUpdateValidatorSharesExchRate{},
	)

	registry.RegisterImplementations((*govtypes.Content)(nil),
		&AddValidatorProposal{},
	)

	// this line is used by starport scaffolding # 3

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
