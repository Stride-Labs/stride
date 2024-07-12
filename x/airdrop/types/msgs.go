package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgClaim         = "claim_daily"
	TypeMsgClaimAndStake = "claim_and_stake"
	TypeMsgClaimEarly    = "claim_early"
)

var (
	_ sdk.Msg = &MsgClaimDaily{}
	_ sdk.Msg = &MsgClaimAndStake{}
	_ sdk.Msg = &MsgClaimEarly{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgClaimDaily{}
	_ legacytx.LegacyMsg = &MsgClaimAndStake{}
	_ legacytx.LegacyMsg = &MsgClaimEarly{}
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
	return TypeMsgClaim
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
