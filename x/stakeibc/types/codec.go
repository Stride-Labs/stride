package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLiquidStake{}, "stakeibc/MsgLiquidStake", nil)
	cdc.RegisterConcrete(&MsgLSMLiquidStake{}, "stakeibc/MsgLSMLiquidStake", nil)
	cdc.RegisterConcrete(&MsgClearBalance{}, "stakeibc/MsgClearBalance", nil)
	cdc.RegisterConcrete(&MsgRegisterHostZone{}, "stakeibc/MsgRegisterHostZone", nil)
	cdc.RegisterConcrete(&MsgRedeemStake{}, "stakeibc/MsgRedeemStake", nil)
	cdc.RegisterConcrete(&MsgClaimUndelegatedTokens{}, "stakeibc/MsgClaimUndelegatedTokens", nil)
	cdc.RegisterConcrete(&MsgRebalanceValidators{}, "stakeibc/MsgRebalanceValidators", nil)
	cdc.RegisterConcrete(&MsgAddValidators{}, "stakeibc/MsgAddValidators", nil)
	cdc.RegisterConcrete(&MsgChangeValidatorWeights{}, "stakeibc/MsgChangeValidatorWeights", nil)
	cdc.RegisterConcrete(&MsgDeleteValidator{}, "stakeibc/MsgDeleteValidator", nil)
	cdc.RegisterConcrete(&AddValidatorsProposal{}, "stakeibc/MsgAddValidatorsProposal", nil)
	cdc.RegisterConcrete(&ToggleLSMProposal{}, "stakeibc/MsgToggleLSMProposal", nil)
	cdc.RegisterConcrete(&MsgRestoreInterchainAccount{}, "stakeibc/MsgRestoreInterchainAccount", nil)
	cdc.RegisterConcrete(&MsgCloseDelegationChannel{}, "stakeibc/MsgCloseDelegationChannel", nil)
	cdc.RegisterConcrete(&MsgUpdateValidatorSharesExchRate{}, "stakeibc/MsgUpdateValidatorSharesExchRate", nil)
	cdc.RegisterConcrete(&MsgCalibrateDelegation{}, "stakeibc/MsgCalibrateDelegation", nil)
	cdc.RegisterConcrete(&MsgCreateTradeRoute{}, "stakeibc/MsgCreateTradeRoute", nil)
	cdc.RegisterConcrete(&MsgDeleteTradeRoute{}, "stakeibc/MsgDeleteTradeRoute", nil)
	cdc.RegisterConcrete(&MsgUpdateTradeRoute{}, "stakeibc/MsgUpdateTradeRoute", nil)
	cdc.RegisterConcrete(&MsgUpdateInnerRedemptionRateBounds{}, "stakeibc/MsgUpdateInnerRedemptionRateBounds", nil)
	cdc.RegisterConcrete(&MsgResumeHostZone{}, "stakeibc/MsgResumeHostZone", nil)
	cdc.RegisterConcrete(&MsgSetCommunityPoolRebate{}, "stakeibc/MsgSetCommunityPoolRebate", nil)
	cdc.RegisterConcrete(&MsgToggleTradeController{}, "stakeibc/MsgToggleTradeController", nil)
	cdc.RegisterConcrete(&MsgUpdateHostZoneParams{}, "stakeibc/MsgUpdateHostZoneParams", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLiquidStake{},
		&MsgClearBalance{},
		&MsgRegisterHostZone{},
		&MsgRedeemStake{},
		&MsgClaimUndelegatedTokens{},
		&MsgRebalanceValidators{},
		&MsgAddValidators{},
		&MsgChangeValidatorWeights{},
		&MsgDeleteValidator{},
		&MsgRestoreInterchainAccount{},
		&MsgCloseDelegationChannel{},
		&MsgUpdateValidatorSharesExchRate{},
		&MsgCalibrateDelegation{},
		&MsgUpdateInnerRedemptionRateBounds{},
		&MsgResumeHostZone{},
		&MsgSetCommunityPoolRebate{},
		&MsgToggleTradeController{},
		&MsgUpdateHostZoneParams{},
	)

	registry.RegisterImplementations((*govtypes.Content)(nil),
		&AddValidatorsProposal{},
		&ToggleLSMProposal{},
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
