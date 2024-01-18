package types

import (
	errorsmod "cosmossdk.io/errors"
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

func NewMsgLiquidStake() *MsgLiquidStake {
	return &MsgLiquidStake{}
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

func NewMsgRedeemStake() *MsgRedeemStake {
	return &MsgRedeemStake{}
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

func NewMsgConfirmDelegation() *MsgConfirmDelegation {
	return &MsgConfirmDelegation{}
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

func NewMsgConfirmUndelegation() *MsgConfirmUndelegation {
	return &MsgConfirmUndelegation{}
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

func NewMsgConfirmUnbondedTokenSweep() *MsgConfirmUnbondedTokenSweep {
	return &MsgConfirmUnbondedTokenSweep{}
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

func NewMsgAdjustDelegatedBalance() *MsgAdjustDelegatedBalance {
	return &MsgAdjustDelegatedBalance{}
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

func NewMsgUpdateInnerRedemptionRateBounds() *MsgUpdateInnerRedemptionRateBounds {
	return &MsgUpdateInnerRedemptionRateBounds{}
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

func NewMsgResumeHostZone() *MsgResumeHostZone {
	return &MsgResumeHostZone{}
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
