package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgLiquidStake{}, "staketia/MsgLiquidStake")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemStake{}, "staketia/MsgRedeemStake")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmDelegation{}, "staketia/MsgConfirmDelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUndelegation{}, "staketia/MsgConfirmUndelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUnbondedTokenSweep{}, "staketia/MsgConfirmUnbondedTokenSweep")
	legacy.RegisterAminoMsg(cdc, &MsgAdjustDelegatedBalance{}, "staketia/MsgAdjustDelegatedBalance")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateInnerRedemptionRateBounds{}, "staketia/MsgUpdateRedemptionRateBounds")
	legacy.RegisterAminoMsg(cdc, &MsgResumeHostZone{}, "staketia/MsgResumeHostZone")
	legacy.RegisterAminoMsg(cdc, &MsgRefreshRedemptionRate{}, "staketia/MsgRefreshRedemptionRate")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteDelegationRecord{}, "staketia/MsgOverwriteDelegationRecord")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteUnbondingRecord{}, "staketia/MsgOverwriteUnbondingRecord")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteRedemptionRecord{}, "staketia/MsgOverwriteRedemptionRecord")
	legacy.RegisterAminoMsg(cdc, &MsgSetOperatorAddress{}, "staketia/MsgSetOperatorAddress")
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
		&MsgRefreshRedemptionRate{},
		&MsgOverwriteDelegationRecord{},
		&MsgOverwriteUnbondingRecord{},
		&MsgOverwriteRedemptionRecord{},
		&MsgSetOperatorAddress{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
