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
	cdc.RegisterConcrete(&MsgLiquidStake{}, "stakeibc/LiquidStake", nil)
	cdc.RegisterConcrete(&MsgLSMLiquidStake{}, "stakeibc/LSMLiquidStake", nil)
	cdc.RegisterConcrete(&MsgRegisterHostZone{}, "stakeibc/RegisterHostZone", nil)
	cdc.RegisterConcrete(&MsgRedeemStake{}, "stakeibc/RedeemStake", nil)
	cdc.RegisterConcrete(&MsgClaimUndelegatedTokens{}, "stakeibc/ClaimUndelegatedTokens", nil)
	cdc.RegisterConcrete(&MsgRebalanceValidators{}, "stakeibc/RebalanceValidators", nil)
	cdc.RegisterConcrete(&MsgAddValidators{}, "stakeibc/AddValidators", nil)
	cdc.RegisterConcrete(&MsgChangeValidatorWeights{}, "stakeibc/ChangeValidatorWeights", nil)
	cdc.RegisterConcrete(&MsgDeleteValidator{}, "stakeibc/DeleteValidator", nil)
	cdc.RegisterConcrete(&AddValidatorsProposal{}, "stakeibc/AddValidatorsProposal", nil)
	cdc.RegisterConcrete(&ToggleLSMProposal{}, "stakeibc/ToggleLSMProposal", nil)
	cdc.RegisterConcrete(&MsgRestoreInterchainAccount{}, "stakeibc/RestoreInterchainAccount", nil)
	cdc.RegisterConcrete(&MsgCloseDelegationChannel{}, "stakeibc/CloseDelegationChannel", nil)
	cdc.RegisterConcrete(&MsgUpdateValidatorSharesExchRate{}, "stakeibc/UpdateValidatorSharesExchRate", nil)
	cdc.RegisterConcrete(&MsgCalibrateDelegation{}, "stakeibc/CalibrateDelegation", nil)
	cdc.RegisterConcrete(&MsgUpdateInnerRedemptionRateBounds{}, "stakeibc/UpdateInnerRedemptionRateBounds", nil)
	cdc.RegisterConcrete(&MsgResumeHostZone{}, "stakeibc/ResumeHostZone", nil)
	cdc.RegisterConcrete(&MsgSetCommunityPoolRebate{}, "stakeibc/SetCommunityPoolRebate", nil)
	cdc.RegisterConcrete(&MsgToggleTradeController{}, "stakeibc/ToggleTradeController", nil)
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
