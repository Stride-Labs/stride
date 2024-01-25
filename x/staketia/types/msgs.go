package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/Stride-Labs/stride/v17/utils"
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
	TypeMsgRefreshRedemptionRate           = "refresh_redemption_rate"
	TypeMsgOverwriteDelegationRecord       = "overwrite_delegation_record"
	TypeMsgOverwriteUnbondingRecord        = "overwrite_unbonding_record"
	TypeMsgOverwriteRedemptionRecord       = "overwrite_redemption_record"
	TypeMsgSetOperatorAddress              = "set_operator_address"
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
	_ sdk.Msg = &MsgRefreshRedemptionRate{}
	_ sdk.Msg = &MsgOverwriteDelegationRecord{}
	_ sdk.Msg = &MsgOverwriteUnbondingRecord{}
	_ sdk.Msg = &MsgOverwriteRedemptionRecord{}
	_ sdk.Msg = &MsgSetOperatorAddress{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgLiquidStake{}
	_ legacytx.LegacyMsg = &MsgRedeemStake{}
	_ legacytx.LegacyMsg = &MsgConfirmDelegation{}
	_ legacytx.LegacyMsg = &MsgConfirmUndelegation{}
	_ legacytx.LegacyMsg = &MsgConfirmUnbondedTokenSweep{}
	_ legacytx.LegacyMsg = &MsgAdjustDelegatedBalance{}
	_ legacytx.LegacyMsg = &MsgUpdateInnerRedemptionRateBounds{}
	_ legacytx.LegacyMsg = &MsgResumeHostZone{}
	_ legacytx.LegacyMsg = &MsgRefreshRedemptionRate{}
	_ legacytx.LegacyMsg = &MsgOverwriteDelegationRecord{}
	_ legacytx.LegacyMsg = &MsgOverwriteUnbondingRecord{}
	_ legacytx.LegacyMsg = &MsgOverwriteRedemptionRecord{}
	_ legacytx.LegacyMsg = &MsgSetOperatorAddress{}
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
	_, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// threshold of 0.1 TIA or 100000 utia avoids denial of service or record spamming
	minThreshold := int64(100000)
	if msg.NativeAmount.LT(sdkmath.NewInt(minThreshold)) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "amount (%v) is below 0.1 TIA minimum", msg.NativeAmount)
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
	_, err := sdk.AccAddressFromBech32(msg.Redeemer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", msg.Redeemer)
	}
	// threshold of 0.1 stTIA or 100000 stutia avoids denial of service or record spamming
	minThreshold := int64(100000)
	if msg.StTokenAmount.LT(sdkmath.NewInt(minThreshold)) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "amount (%v) is below 0.1 stTIA minimum", msg.StTokenAmount)
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
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// Note: We can't verify admin in ValidateBasic, because it requires inspecting the HostZone
	// Note: We can't verify recordId in ValidateBasic, because 0 is a valid record id
	// and recordId is uint64 so can't be negative
	if err := utils.VerifyTxHash(msg.TxHash); err != nil {
		return err
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
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// verify tx hash is valid
	if err := utils.VerifyTxHash(msg.TxHash); err != nil {
		return err
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
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// Note: We can't verify this is sent by the safe or operator in ValidateBasic, because it requires inspecting the HostZone
	// Note: We can't verify recordId in ValidateBasic, because 0 is a valid record id
	// and recordId is uint64 so can't be negative
	if err := utils.VerifyTxHash(msg.TxHash); err != nil {
		return err
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
	_, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.DelegationOffset.IsNil() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "delegation offset must be specified")
	}
	if msg.ValidatorAddress == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "validator address must be specified")
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
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// Confirm the max is greater than the min
	if msg.MaxInnerRedemptionRate.LTE(msg.MinInnerRedemptionRate) {
		return errorsmod.Wrapf(ErrInvalidRedemptionRateBounds,
			"Inner max safety threshold (%s) is less than inner min safety threshold (%s)",
			msg.MaxInnerRedemptionRate, msg.MinInnerRedemptionRate)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
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
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------
//             MsgSetOperatorAddress
// ----------------------------------------------

func NewMsgSetOperatorAddress(signer string, operator string) *MsgSetOperatorAddress {
	return &MsgSetOperatorAddress{
		Signer:   signer,
		Operator: operator,
	}
}

func (msg MsgSetOperatorAddress) Type() string {
	return TypeMsgSetOperatorAddress
}

func (msg MsgSetOperatorAddress) Route() string {
	return RouterKey
}

func (msg *MsgSetOperatorAddress) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetOperatorAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetOperatorAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//       MsgRefreshRedemptionRate
// ----------------------------------------------

func NewMsgRefreshRedemptionRate(creator string) *MsgRefreshRedemptionRate {
	return &MsgRefreshRedemptionRate{
		Creator: creator,
	}
}

func (msg MsgRefreshRedemptionRate) Type() string {
	return TypeMsgRefreshRedemptionRate
}

func (msg MsgRefreshRedemptionRate) Route() string {
	return RouterKey
}

func (msg *MsgRefreshRedemptionRate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRefreshRedemptionRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRefreshRedemptionRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	return nil
}

// ----------------------------------------------
//       MsgOverwriteDelegationRecord
// ----------------------------------------------

func NewMsgOverwriteDelegationRecord(creator string, delegationRecord DelegationRecord) *MsgOverwriteDelegationRecord {
	return &MsgOverwriteDelegationRecord{
		Creator:          creator,
		DelegationRecord: &delegationRecord,
	}
}

func (msg MsgOverwriteDelegationRecord) Type() string {
	return TypeMsgOverwriteDelegationRecord
}

func (msg MsgOverwriteDelegationRecord) Route() string {
	return RouterKey
}

func (msg *MsgOverwriteDelegationRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgOverwriteDelegationRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgOverwriteDelegationRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	// Check the record's attributes
	// - assert the nativeAmount is non-negative (zero is acceptable)
	if msg.DelegationRecord.NativeAmount.LT(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "amount < 0")
	}

	// - assert the status is one of the acceptable statuses
	if _, ok := DelegationRecordStatus_name[int32(msg.DelegationRecord.Status)]; !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "record status doesn't match the enum")
	}

	return nil
}

// ----------------------------------------------
//       MsgOverwriteUnbondingRecord
// ----------------------------------------------

func NewMsgOverwriteUnbondingRecord(creator string, unbondingRecord UnbondingRecord) *MsgOverwriteUnbondingRecord {
	return &MsgOverwriteUnbondingRecord{
		Creator:         creator,
		UnbondingRecord: &unbondingRecord,
	}
}

func (msg MsgOverwriteUnbondingRecord) Type() string {
	return TypeMsgOverwriteUnbondingRecord
}

func (msg MsgOverwriteUnbondingRecord) Route() string {
	return RouterKey
}

func (msg *MsgOverwriteUnbondingRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgOverwriteUnbondingRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgOverwriteUnbondingRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	// Check the record's attributes
	// - assert the nativeAmount is non-negative (zero is acceptable)
	if msg.UnbondingRecord.NativeAmount.LT(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "native amount < 0")
	}
	// - assert the stTokenAmount is non-negative (zero is acceptable)
	if msg.UnbondingRecord.StTokenAmount.LT(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "sttoken amount < 0")
	}

	// - assert the status is one of the acceptable statuses
	if _, ok := UnbondingRecordStatus_name[int32(msg.UnbondingRecord.Status)]; !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "record status doesn't match the enum")
	}

	return nil
}

// ----------------------------------------------
//       MsgOverwriteRedemptionRecord
// ----------------------------------------------

func NewMsgOverwriteRedemptionRecord(creator string, redemptionRecord RedemptionRecord) *MsgOverwriteRedemptionRecord {
	return &MsgOverwriteRedemptionRecord{
		Creator:          creator,
		RedemptionRecord: &redemptionRecord,
	}
}

func (msg MsgOverwriteRedemptionRecord) Type() string {
	return TypeMsgOverwriteRedemptionRecord
}

func (msg MsgOverwriteRedemptionRecord) Route() string {
	return RouterKey
}

func (msg *MsgOverwriteRedemptionRecord) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgOverwriteRedemptionRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgOverwriteRedemptionRecord) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	// Check the record's attributes
	// - assert the redeemer is a valid address
	_, err = sdk.AccAddressFromBech32(msg.RedemptionRecord.Redeemer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	// - assert the nativeAmount is non-negative (zero is acceptable)
	if msg.RedemptionRecord.NativeAmount.LT(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "amount < 0")
	}
	// - assert the stTokenAmount is non-negative (zero is acceptable)
	if msg.RedemptionRecord.StTokenAmount.LT(sdk.ZeroInt()) {
		return errorsmod.Wrapf(ErrInvalidAmountBelowMinimum, "amount < 0")
	}

	return nil
}
