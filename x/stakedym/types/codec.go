package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgLiquidStake{}, "stakedym/MsgLiquidStake")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemStake{}, "stakedym/MsgRedeemStake")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmDelegation{}, "stakedym/MsgConfirmDelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUndelegation{}, "stakedym/MsgConfirmUndelegation")
	legacy.RegisterAminoMsg(cdc, &MsgConfirmUnbondedTokenSweep{}, "stakedym/MsgConfirmUnbondedTokenSweep")
	legacy.RegisterAminoMsg(cdc, &MsgAdjustDelegatedBalance{}, "stakedym/MsgAdjustDelegatedBalance")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateInnerRedemptionRateBounds{}, "stakedym/MsgUpdateRedemptionRateBounds")
	legacy.RegisterAminoMsg(cdc, &MsgResumeHostZone{}, "stakedym/MsgResumeHostZone")
	legacy.RegisterAminoMsg(cdc, &MsgRefreshRedemptionRate{}, "stakedym/MsgRefreshRedemptionRate")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteDelegationRecord{}, "stakedym/MsgOverwriteDelegationRecord")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteUnbondingRecord{}, "stakedym/MsgOverwriteUnbondingRecord")
	legacy.RegisterAminoMsg(cdc, &MsgOverwriteRedemptionRecord{}, "stakedym/MsgOverwriteRedemptionRecord")
	legacy.RegisterAminoMsg(cdc, &MsgSetOperatorAddress{}, "stakedym/MsgSetOperatorAddress")
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
