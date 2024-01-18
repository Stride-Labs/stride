package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgLiquidStake                     = "liquid_stake"
	TypeMsgRedeemStake                     = "redeem_stake"
	TypeMsgConfirmDelegation               = "confirm_delegation"
	TypeMsgConfirmUndelegation             = "confirm_undelegation"
	TypeMsgConfirmUnbondedTokenSweep       = "confirm_unbonded_token_sweep"
	TypeMsgAdjustDelegatedBalance          = "adjust_delegated_balance"
	TypeMsgUpdateInnerRedemptionRateBounds = "redemption_rate_bounds"
	TypeMsgResumeHostZone                  = "resume_host_zone"
)

var (
	_ sdk.Msg = &MsgLiquidStake{}
	_ sdk.Msg = &MsgRedeemStake{}
	_ sdk.Msg = &MsgConfirmDelegation{}
	_ sdk.Msg = &MsgConfirmUndelegation{}
	_ sdk.Msg = &MsgConfirmUnbondedTokenSweep{}
	_ sdk.Msg = &MsgAdjustDelegatedBalance{}
	_ sdk.Msg = &MsgUpdateInnerRedemptionRateBounds{}
	_ sdk.Msg = &MsgResumeHostZone{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgLiquidStake{}
	_ legacytx.LegacyMsg = &MsgRedeemStake{}
	_ legacytx.LegacyMsg = &MsgConfirmDelegation{}
	_ legacytx.LegacyMsg = &MsgConfirmUndelegation{}
	_ legacytx.LegacyMsg = &MsgConfirmUnbondedTokenSweep{}
	_ legacytx.LegacyMsg = &MsgAdjustDelegatedBalance{}
	_ legacytx.LegacyMsg = &MsgUpdateInnerRedemptionRateBounds{}
	_ legacytx.LegacyMsg = &MsgResumeHostZone{}
)

// ----------------------------------------------
//               MsgLiquidStake
// ----------------------------------------------

func NewMsgLiquidStake(staker string, nativeAmount sdkmath.Int) *MsgLiquidStake {
	return &MsgLiquidStake{
		Staker:       staker,
		NativeAmount: nativeAmount,
	}
}

func (msg MsgLiquidStake) Type() string {
	return TypeMsgLiquidStake
}

func (msg MsgLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLiquidStake) GetSigners() []sdk.AccAddress {
	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{staker}
}

func (msg *MsgLiquidStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLiquidStake) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//               MsgRedeemStake
// ----------------------------------------------

func NewMsgRedeemStake(redeemer string, stTokenAmount sdkmath.Int) *MsgRedeemStake {
	return &MsgRedeemStake{
		Redeemer:      redeemer,
		StTokenAmount: stTokenAmount,
	}
}

func (msg MsgRedeemStake) Type() string {
	return TypeMsgRedeemStake
}

func (msg MsgRedeemStake) Route() string {
	return RouterKey
}

func (msg *MsgRedeemStake) GetSigners() []sdk.AccAddress {
	redeemer, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{redeemer}
}

func (msg *MsgRedeemStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemStake) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//             MsgConfirmDelegation
// ----------------------------------------------

func NewMsgConfirmDelegation(operator string, recordId uint64, txHash string) *MsgConfirmDelegation {
	return &MsgConfirmDelegation{
		Operator: operator,
		RecordId: recordId,
		TxHash:   txHash,
	}
}

func (msg MsgConfirmDelegation) Type() string {
	return TypeMsgConfirmDelegation
}

func (msg MsgConfirmDelegation) Route() string {
	return RouterKey
}

func (msg *MsgConfirmDelegation) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgConfirmDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfirmDelegation) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//            MsgConfirmUndelegation
// ----------------------------------------------

func NewMsgConfirmUndelegation(operator string, recordId uint64, txHash string) *MsgConfirmUndelegation {
	return &MsgConfirmUndelegation{
		Operator: operator,
		RecordId: recordId,
		TxHash:   txHash,
	}
}

func (msg MsgConfirmUndelegation) Type() string {
	return TypeMsgConfirmUndelegation
}

func (msg MsgConfirmUndelegation) Route() string {
	return RouterKey
}

func (msg *MsgConfirmUndelegation) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgConfirmUndelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfirmUndelegation) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//          MsgConfirmUnbondedTokenSweep
// ----------------------------------------------

func NewMsgConfirmUnbondedTokenSweep(operator string, recordId uint64, txHash string) *MsgConfirmUnbondedTokenSweep {
	return &MsgConfirmUnbondedTokenSweep{
		Operator: operator,
		RecordId: recordId,
		TxHash:   txHash,
	}
}

func (msg MsgConfirmUnbondedTokenSweep) Type() string {
	return TypeMsgConfirmUnbondedTokenSweep
}

func (msg MsgConfirmUnbondedTokenSweep) Route() string {
	return RouterKey
}

func (msg *MsgConfirmUnbondedTokenSweep) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgConfirmUnbondedTokenSweep) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfirmUnbondedTokenSweep) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//          MsgAdjustDelegatedBalance
// ----------------------------------------------

func NewMsgAdjustDelegatedBalance(operator string, delegationOffset sdkmath.Int, validatorAddress string) *MsgAdjustDelegatedBalance {
	return &MsgAdjustDelegatedBalance{
		Operator:         operator,
		DelegationOffset: delegationOffset,
		ValidatorAddress: validatorAddress,
	}
}

func (msg MsgAdjustDelegatedBalance) Type() string {
	return TypeMsgAdjustDelegatedBalance
}

func (msg MsgAdjustDelegatedBalance) Route() string {
	return RouterKey
}

func (msg *MsgAdjustDelegatedBalance) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

func (msg *MsgAdjustDelegatedBalance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAdjustDelegatedBalance) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//       MsgUpdateInnerRedemptionRateBounds
// ----------------------------------------------

func NewMsgUpdateInnerRedemptionRateBounds(creator string, minRedemptionRate, maxRedemptionRate sdk.Dec) *MsgUpdateInnerRedemptionRateBounds {
	return &MsgUpdateInnerRedemptionRateBounds{
		Creator:                creator,
		MinInnerRedemptionRate: minRedemptionRate,
		MaxInnerRedemptionRate: maxRedemptionRate,
	}
}

func (msg MsgUpdateInnerRedemptionRateBounds) Type() string {
	return TypeMsgUpdateInnerRedemptionRateBounds
}

func (msg MsgUpdateInnerRedemptionRateBounds) Route() string {
	return RouterKey
}

func (msg *MsgUpdateInnerRedemptionRateBounds) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateInnerRedemptionRateBounds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateInnerRedemptionRateBounds) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//             MsgResumeHostZone
// ----------------------------------------------

func NewMsgResumeHostZone(creator string) *MsgResumeHostZone {
	return &MsgResumeHostZone{
		Creator: creator,
	}
}

func (msg MsgResumeHostZone) Type() string {
	return TypeMsgResumeHostZone
}

func (msg MsgResumeHostZone) Route() string {
	return RouterKey
}

func (msg *MsgResumeHostZone) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgResumeHostZone) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResumeHostZone) ValidateBasic() error {
	// TODO [sttia]
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}
