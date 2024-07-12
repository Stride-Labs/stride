package types

import (
	"errors"
	fmt "fmt"
	time "time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	"github.com/Stride-Labs/stride/v22/utils"
)

const (
	TypeMsgClaimDaily    = "claim_daily"
	TypeMsgClaimAndStake = "claim_and_stake"
	TypeMsgClaimEarly    = "claim_early"

	TypeMsgCreateAirdrop        = "create_airdrop"
	TypeMsgUpdateAirdrop        = "update_airdrop"
	TypeMsgAddAllocations       = "add_allocations"
	TypeMsgUpdateUserAllocation = "update_user_allocation"
	TypeMsgLinkAddresses        = "link_addresses"
)

var (
	_ sdk.Msg = &MsgClaimDaily{}
	_ sdk.Msg = &MsgClaimAndStake{}
	_ sdk.Msg = &MsgClaimEarly{}

	_ sdk.Msg = &MsgCreateAirdrop{}
	_ sdk.Msg = &MsgUpdateAirdrop{}
	_ sdk.Msg = &MsgAddAllocations{}
	_ sdk.Msg = &MsgUpdateUserAllocation{}
	_ sdk.Msg = &MsgLinkAddresses{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgClaimDaily{}
	_ legacytx.LegacyMsg = &MsgClaimAndStake{}
	_ legacytx.LegacyMsg = &MsgClaimEarly{}

	_ legacytx.LegacyMsg = &MsgCreateAirdrop{}
	_ legacytx.LegacyMsg = &MsgUpdateAirdrop{}
	_ legacytx.LegacyMsg = &MsgAddAllocations{}
	_ legacytx.LegacyMsg = &MsgUpdateUserAllocation{}
	_ legacytx.LegacyMsg = &MsgLinkAddresses{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgClaimDaily(claimer, airdropId string) *MsgClaimDaily {
	return &MsgClaimDaily{
		Claimer:   claimer,
		AirdropId: airdropId,
	}
}

func (msg MsgClaimDaily) Type() string {
	return TypeMsgClaimDaily
}

func (msg MsgClaimDaily) Route() string {
	return RouterKey
}

func (msg *MsgClaimDaily) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgClaimDaily) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimDaily) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Claimer); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}

	return nil
}

// ----------------------------------------------
//               MsgClaimEarly
// ----------------------------------------------

func NewMsgClaimEarly(claimer, airdropId string) *MsgClaimEarly {
	return &MsgClaimEarly{
		Claimer:   claimer,
		AirdropId: airdropId,
	}
}

func (msg MsgClaimEarly) Type() string {
	return TypeMsgClaimEarly
}

func (msg MsgClaimEarly) Route() string {
	return RouterKey
}

func (msg *MsgClaimEarly) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgClaimEarly) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimEarly) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Claimer); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}

	return nil
}

// ----------------------------------------------
//               MsgClaimAndStake
// ----------------------------------------------

func NewMsgClaimAndStake(claimer, airdropId, validatorAddress string) *MsgClaimAndStake {
	return &MsgClaimAndStake{
		Claimer:          claimer,
		AirdropId:        airdropId,
		ValidatorAddress: validatorAddress,
	}
}

func (msg MsgClaimAndStake) Type() string {
	return TypeMsgClaimAndStake
}

func (msg MsgClaimAndStake) Route() string {
	return RouterKey
}

func (msg *MsgClaimAndStake) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgClaimAndStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimAndStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Claimer); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgCreateAirdrop
// ----------------------------------------------

func NewMsgCreateAirdrop(
	admin string,
	airdropId string,
	distributionStartDate *time.Time,
	distributionEndDate *time.Time,
	clawbackDate *time.Time,
	claimDeadlineDate *time.Time,
	earlyClaimPenalty sdk.Dec,
	claimAndStakeBonus sdk.Dec,
	distributionAddress string,
) *MsgCreateAirdrop {
	return &MsgCreateAirdrop{
		Admin:                 admin,
		AirdropId:             airdropId,
		DistributionStartDate: distributionStartDate,
		DistributionEndDate:   distributionEndDate,
		ClawbackDate:          clawbackDate,
		ClaimTypeDeadlineDate: claimDeadlineDate,
		EarlyClaimPenalty:     earlyClaimPenalty,
		ClaimAndStakeBonus:    claimAndStakeBonus,
		DistributionAddress:   distributionAddress,
	}
}

func (msg MsgCreateAirdrop) Type() string {
	return TypeMsgCreateAirdrop
}

func (msg MsgCreateAirdrop) Route() string {
	return RouterKey
}

func (msg *MsgCreateAirdrop) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgCreateAirdrop) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateAirdrop) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}

	if msg.DistributionStartDate == nil || *msg.DistributionStartDate == (time.Time{}) {
		return errors.New("distribution start date must be specified")
	}
	if msg.DistributionEndDate == nil || *msg.DistributionEndDate == (time.Time{}) {
		return errors.New("distribution end date must be specified")
	}
	if msg.ClawbackDate == nil || *msg.ClawbackDate == (time.Time{}) {
		return errors.New("clawback date must be specified")
	}
	if msg.ClaimTypeDeadlineDate == nil || *msg.ClaimTypeDeadlineDate == (time.Time{}) {
		return errors.New("claim type deadline date must be specified")
	}

	if !msg.DistributionEndDate.After(*msg.DistributionStartDate) {
		return errors.New("distribution end date must be after the start date")
	}
	if !msg.ClaimTypeDeadlineDate.After(*msg.DistributionStartDate) {
		return errors.New("claim type deadline date must be after the distribution start date")
	}
	if !msg.ClaimTypeDeadlineDate.Before(*msg.DistributionEndDate) {
		return errors.New("claim type deadline date must be before the distribution end date")
	}
	if !msg.ClawbackDate.After(*msg.DistributionEndDate) {
		return errors.New("clawback date must be after the distribution end date")
	}

	if msg.EarlyClaimPenalty.IsNil() {
		return errors.New("early claim penalty must be specified")
	}
	if msg.ClaimAndStakeBonus.IsNil() {
		return errors.New("claim and stake bonus must be specified")
	}

	if msg.EarlyClaimPenalty.LT(sdk.ZeroDec()) || msg.EarlyClaimPenalty.GT(sdk.OneDec()) {
		return errors.New("early claim penalty must be between 0 and 1")
	}
	if msg.ClaimAndStakeBonus.LT(sdk.ZeroDec()) || msg.ClaimAndStakeBonus.GT(sdk.OneDec()) {
		return errors.New("early claim penalty must be between 0 and 1")
	}

	if _, err := sdk.AccAddressFromBech32(msg.DistributionAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distribution address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgUpdateAirdrop
// ----------------------------------------------

func NewMsgUpdateAirdrop(
	admin string,
	airdropId string,
	distributionStartDate *time.Time,
	distributionEndDate *time.Time,
	clawbackDate *time.Time,
	claimDeadlineDate *time.Time,
	earlyClaimPenalty sdk.Dec,
	claimAndStakeBonus sdk.Dec,
	distributionAddress string,
) *MsgUpdateAirdrop {
	return &MsgUpdateAirdrop{
		Admin:                 admin,
		AirdropId:             airdropId,
		DistributionStartDate: distributionStartDate,
		DistributionEndDate:   distributionEndDate,
		ClawbackDate:          clawbackDate,
		ClaimTypeDeadlineDate: claimDeadlineDate,
		EarlyClaimPenalty:     earlyClaimPenalty,
		ClaimAndStakeBonus:    claimAndStakeBonus,
		DistributionAddress:   distributionAddress,
	}
}

func (msg MsgUpdateAirdrop) Type() string {
	return TypeMsgUpdateAirdrop
}

func (msg MsgUpdateAirdrop) Route() string {
	return RouterKey
}

func (msg *MsgUpdateAirdrop) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgUpdateAirdrop) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateAirdrop) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}

	if msg.DistributionStartDate == nil || *msg.DistributionStartDate == (time.Time{}) {
		return errors.New("distribution start date must be specified")
	}
	if msg.DistributionEndDate == nil || *msg.DistributionEndDate == (time.Time{}) {
		return errors.New("distribution end date must be specified")
	}
	if msg.ClawbackDate == nil || *msg.ClawbackDate == (time.Time{}) {
		return errors.New("clawback date must be specified")
	}
	if msg.ClaimTypeDeadlineDate == nil || *msg.ClaimTypeDeadlineDate == (time.Time{}) {
		return errors.New("claim type deadline date must be specified")
	}

	if !msg.DistributionEndDate.After(*msg.DistributionStartDate) {
		return errors.New("distribution end date must be after the start date")
	}
	if !msg.ClaimTypeDeadlineDate.After(*msg.DistributionStartDate) {
		return errors.New("claim type deadline date must be after the distribution start date")
	}
	if !msg.ClaimTypeDeadlineDate.Before(*msg.DistributionEndDate) {
		return errors.New("claim type deadline date must be before the distribution end date")
	}
	if !msg.ClawbackDate.After(*msg.DistributionEndDate) {
		return errors.New("clawback date must be after the distribution end date")
	}

	if msg.EarlyClaimPenalty.IsNil() {
		return errors.New("early claim penalty must be specified")
	}
	if msg.ClaimAndStakeBonus.IsNil() {
		return errors.New("claim and stake bonus must be specified")
	}

	if msg.EarlyClaimPenalty.LT(sdk.ZeroDec()) || msg.EarlyClaimPenalty.GT(sdk.OneDec()) {
		return errors.New("early claim penalty must be between 0 and 1")
	}
	if msg.ClaimAndStakeBonus.LT(sdk.ZeroDec()) || msg.ClaimAndStakeBonus.GT(sdk.OneDec()) {
		return errors.New("early claim penalty must be between 0 and 1")
	}

	if _, err := sdk.AccAddressFromBech32(msg.DistributionAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distribution address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//             MsgAddAllocations
// ----------------------------------------------

func NewMsgAddAllocations(admin string, airdropId string, allocations []RawAllocation) *MsgAddAllocations {
	return &MsgAddAllocations{
		Admin:       admin,
		AirdropId:   airdropId,
		Allocations: allocations,
	}
}

func (msg MsgAddAllocations) Type() string {
	return TypeMsgAddAllocations
}

func (msg MsgAddAllocations) Route() string {
	return RouterKey
}

func (msg *MsgAddAllocations) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgAddAllocations) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddAllocations) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}

	addresses := map[string]bool{}
	for _, allocation := range msg.Allocations {
		if allocation.UserAddress == "" {
			return errors.New("all addresses in allocations must be specified")
		}

		if _, ok := addresses[allocation.UserAddress]; ok {
			return fmt.Errorf("address %s is specified more than once", allocation.UserAddress)
		}
		addresses[allocation.UserAddress] = true

		for _, amount := range allocation.Allocations {
			if amount.IsNil() || amount.LT(sdkmath.ZeroInt()) {
				return errors.New("all allocation amounts must be specified and positive")
			}
		}
	}

	return nil
}

// ----------------------------------------------
//             MsgUpdateUserAllocation
// ----------------------------------------------

func NewMsgUpdateUserAllocation(admin, airdropId, userAddress string, allocations []sdkmath.Int) *MsgUpdateUserAllocation {
	return &MsgUpdateUserAllocation{
		Admin:       admin,
		AirdropId:   airdropId,
		UserAddress: userAddress,
		Allocations: allocations,
	}
}

func (msg MsgUpdateUserAllocation) Type() string {
	return TypeMsgUpdateUserAllocation
}

func (msg MsgUpdateUserAllocation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateUserAllocation) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgUpdateUserAllocation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateUserAllocation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}
	if msg.UserAddress == "" {
		return errors.New("user address must be specified")
	}

	for _, allocation := range msg.Allocations {
		if allocation.IsNil() || allocation.LT(sdk.ZeroInt()) {
			return errors.New("all allocation amounts must be specified and positive")
		}
	}

	return nil
}

// ----------------------------------------------
//             MsgLinkAddresses
// ----------------------------------------------

func NewMsgLinkAddresses(admin, airdropId, strideAddress, hostAddress string) *MsgLinkAddresses {
	return &MsgLinkAddresses{
		Admin:         admin,
		AirdropId:     airdropId,
		StrideAddress: strideAddress,
		HostAddress:   hostAddress,
	}
}

func (msg MsgLinkAddresses) Type() string {
	return TypeMsgLinkAddresses
}

func (msg MsgLinkAddresses) Route() string {
	return RouterKey
}

func (msg *MsgLinkAddresses) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgLinkAddresses) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLinkAddresses) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.StrideAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid stride address (%s)", err)
	}
	if msg.AirdropId == "" {
		return errors.New("airdrop-id must be specified")
	}
	if msg.HostAddress == "" {
		return errors.New("host address must be specified")
	}

	return nil
}
