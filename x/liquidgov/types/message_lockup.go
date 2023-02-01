package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

const (
	TypeMsgLockupTokens = "lockup_tokens"
	TypeMsgUnlockTokens = "begin_unlocking"
)

var (
	_ sdk.Msg = &MsgLockupTokens{}
	_ sdk.Msg = &MsgUnlockTokens{}
)

func NewMsgLockupTokens(creator string, amount sdkmath.Int, denom string) *MsgLockupTokens {
	return &MsgLockupTokens{
		Creator: creator,
		Amount:  amount,
		Denom:   denom,
	}
}

func (msg *MsgLockupTokens) Route() string {
	return RouterKey
}

func (msg *MsgLockupTokens) Type() string {
	return TypeMsgLockupTokens
}

func (msg *MsgLockupTokens) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLockupTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLockupTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate amount is positive nonzero
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return sdkerrors.Wrapf(stakeibctypes.ErrInvalidAmount, "amount locked up must be positive and nonzero")
	}
	// validate host denom is not empty
	if msg.Denom == "" {
		return sdkerrors.Wrapf(stakeibctypes.ErrRequiredFieldEmpty, "denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}
	return nil
}

// NewMsgUndelegate creates a new MsgUndelegate instance.
//
//nolint:interfacer
func NewMsgUnlockTokens(delAddr sdk.AccAddress, denom string, amount sdk.Int) *MsgUnlockTokens {
	return &MsgUnlockTokens{
		Creator: delAddr.String(),
		Denom:   denom,
		Amount:  amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgUnlockTokens) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgUnlockTokens) Type() string { return TypeMsgUnlockTokens }

// GetSigners implements the sdk.Msg interface.
func (msg MsgUnlockTokens) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// GetSignBytes implements the sdk.Msg interface.
func (msg MsgUnlockTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgUnlockTokens) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid creator address: %s", err)
	}
	// validate amount is positive nonzero
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return sdkerrors.Wrapf(stakeibctypes.ErrInvalidAmount, "amount unlocked must be positive and nonzero")
	}
	// validate host denom is not empty
	if msg.Denom == "" {
		return sdkerrors.Wrapf(stakeibctypes.ErrRequiredFieldEmpty, "denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return err
	}

	return nil
}
