package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgLiquidStake{}, "stakeibc/MsgLiquidStake")
	legacy.RegisterAminoMsg(cdc, &MsgLSMLiquidStake{}, "stakeibc/MsgLSMLiquidStake")
	legacy.RegisterAminoMsg(cdc, &MsgClearBalance{}, "stakeibc/MsgClearBalance")
	legacy.RegisterAminoMsg(cdc, &MsgRegisterHostZone{}, "stakeibc/MsgRegisterHostZone")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemStake{}, "stakeibc/MsgRedeemStake")
	legacy.RegisterAminoMsg(cdc, &MsgClaimUndelegatedTokens{}, "stakeibc/MsgClaimUndelegatedTokens")
	legacy.RegisterAminoMsg(cdc, &MsgRebalanceValidators{}, "stakeibc/MsgRebalanceValidators")
	legacy.RegisterAminoMsg(cdc, &MsgAddValidators{}, "stakeibc/MsgAddValidators")
	legacy.RegisterAminoMsg(cdc, &MsgChangeValidatorWeights{}, "stakeibc/MsgChangeValidatorWeights")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteValidator{}, "stakeibc/MsgDeleteValidator")
	cdc.RegisterConcrete(&AddValidatorsProposal{}, "stakeibc/AddValidatorsProposal", nil)
	cdc.RegisterConcrete(&ToggleLSMProposal{}, "stakeibc/ToggleLSMProposal", nil)
	legacy.RegisterAminoMsg(cdc, &MsgRestoreInterchainAccount{}, "stakeibc/MsgRestoreInterchainAccount")
	legacy.RegisterAminoMsg(cdc, &MsgCloseDelegationChannel{}, "stakeibc/MsgCloseDelegationChannel")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateValSharesExchRate{}, "stakeibc/MsgUpdateValSharesExchRate")
	legacy.RegisterAminoMsg(cdc, &MsgCalibrateDelegation{}, "stakeibc/MsgCalibrateDelegation")
	legacy.RegisterAminoMsg(cdc, &MsgCreateTradeRoute{}, "stakeibc/MsgCreateTradeRoute")
	legacy.RegisterAminoMsg(cdc, &MsgDeleteTradeRoute{}, "stakeibc/MsgDeleteTradeRoute")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateTradeRoute{}, "stakeibc/MsgUpdateTradeRoute")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateRedemptionRateBounds{}, "stakeibc/MsgUpdateRedemptionRateBounds")
	legacy.RegisterAminoMsg(cdc, &MsgResumeHostZone{}, "stakeibc/MsgResumeHostZone")
	legacy.RegisterAminoMsg(cdc, &MsgSetCommunityPoolRebate{}, "stakeibc/MsgSetCommunityPoolRebate")
	legacy.RegisterAminoMsg(cdc, &MsgToggleTradeController{}, "stakeibc/MsgToggleTradeController")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateHostZoneParams{}, "stakeibc/MsgUpdateHostZoneParams")
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
		&MsgUpdateValSharesExchRate{},
		&MsgCalibrateDelegation{},
		&MsgUpdateRedemptionRateBounds{},
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
