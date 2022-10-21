package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Msg type for MsgSetAirdropAllocations
const TypeMsgSetAirdropAllocations = "set_airdrop_allocation"

var _ sdk.Msg = &MsgSetAirdropAllocations{}

func NewMsgSetAirdropAllocations(allocator string, airdropIdentifier string, users []string, weights []sdk.Dec) *MsgSetAirdropAllocations {
	return &MsgSetAirdropAllocations{
		Allocator:         allocator,
		AirdropIdentifier: airdropIdentifier,
		Users:             users,
		Weights:           weights,
	}
}

func (msg *MsgSetAirdropAllocations) Route() string {
	return RouterKey
}

func (msg *MsgSetAirdropAllocations) Type() string {
	return TypeMsgSetAirdropAllocations
}

func (msg *MsgSetAirdropAllocations) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.Allocator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
}

func (msg *MsgSetAirdropAllocations) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetAirdropAllocations) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Allocator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid allocator address (%s)", err)
	}

	if msg.AirdropIdentifier == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop identifier not set")
	}

	if len(msg.Users) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty users list")
	}

	if len(msg.Weights) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "empty weights list")
	}

	if len(msg.Users) != len(msg.Weights) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "different length")
	}

	for _, user := range msg.Users {
		_, err := sdk.AccAddressFromBech32(user)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid user address (%s)", err)
		}
	}

	for _, weight := range msg.Weights {
		if weight.Equal(sdk.NewDec(0)) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid user weight")
		}
	}

	return nil
}

// Msg type for MsgClaimFreeAmount
const TypeMsgClaimFreeAmount = "claim_free_amount"

var _ sdk.Msg = &MsgClaimFreeAmount{}

func NewMsgClaimFreeAmount(user string, airdropIdentifier string) *MsgClaimFreeAmount {
	return &MsgClaimFreeAmount{
		User:              user,
		AirdropIdentifier: airdropIdentifier,
	}
}

func (msg *MsgClaimFreeAmount) Route() string {
	return RouterKey
}

func (msg *MsgClaimFreeAmount) Type() string {
	return TypeMsgClaimFreeAmount
}

func (msg *MsgClaimFreeAmount) GetSigners() []sdk.AccAddress {
	allocator, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{allocator}
}

func (msg *MsgClaimFreeAmount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimFreeAmount) ValidateBasic() error {
	if msg.AirdropIdentifier == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop identifier not set")
	}

	_, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid user address (%s)", err)
	}

	return nil
}
