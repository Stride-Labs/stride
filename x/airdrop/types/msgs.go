package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgClaim         = "claim"
	TypeMsgClaimAndStake = "claim_and_stake"
	TypeMsgClaimEarly    = "claim_early"
)

var (
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgClaimAndStake{}
	_ sdk.Msg = &MsgClaimEarly{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgClaim{}
	_ legacytx.LegacyMsg = &MsgClaimAndStake{}
	_ legacytx.LegacyMsg = &MsgClaimEarly{}
)

// ----------------------------------------------
//               MsgClaim
// ----------------------------------------------

func NewMsgClaim(claimer string) *MsgClaim {
	return &MsgClaim{
		Claimer: claimer,
	}
}

func (msg MsgClaim) Type() string {
	return TypeMsgClaim
}

func (msg MsgClaim) Route() string {
	return RouterKey
}

func (msg *MsgClaim) GetSigners() []sdk.AccAddress {
	claimer, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{claimer}
}

func (msg *MsgClaim) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaim) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgClaimAndStake
// ----------------------------------------------

func NewMsgClaimAndStake(claimer string) *MsgClaimAndStake {
	return &MsgClaimAndStake{
		Claimer: claimer,
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
	_, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgClaimEarly
// ----------------------------------------------

func NewMsgClaimEarly(claimer string) *MsgClaimEarly {
	return &MsgClaimEarly{
		Claimer: claimer,
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
	_, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}
