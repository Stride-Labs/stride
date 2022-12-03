package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/utils"
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
		strideAddr := utils.ConvertAddressToStrideAddress(user)
		if strideAddr == "" {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid bech32 address")
		}

		_, err := sdk.AccAddressFromBech32(strideAddr)
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

func NewMsgClaimFreeAmount(user string) *MsgClaimFreeAmount {
	return &MsgClaimFreeAmount{
		User: user,
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

	_, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid user address (%s)", err)
	}

	return nil
}

// Msg type for MsgCreateAirdrop
const TypeMsgCreateAirdrop = "create_airdrop"

var _ sdk.Msg = &MsgCreateAirdrop{}

func NewMsgCreateAirdrop(distributor string, identifier string, startTime uint64, duration uint64, denom string) *MsgCreateAirdrop {
	return &MsgCreateAirdrop{
		Distributor: distributor,
		Identifier:  identifier,
		StartTime:   startTime,
		Duration:    duration,
		Denom:       denom,
	}
}

func (msg *MsgCreateAirdrop) Route() string {
	return RouterKey
}

func (msg *MsgCreateAirdrop) Type() string {
	return TypeMsgCreateAirdrop
}

func (msg *MsgCreateAirdrop) GetSigners() []sdk.AccAddress {
	distributor, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{distributor}
}

func (msg *MsgCreateAirdrop) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateAirdrop) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address (%s)", err)
	}

	if msg.Identifier == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop identifier not set")
	}

	if msg.StartTime == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop start time not set")
	}

	if msg.Duration == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop duration not set")
	}
	return nil
}

// Msg type for MsgDeleteAirdrop
const TypeMsgDeleteAirdrop = "delete_airdrop"

var _ sdk.Msg = &MsgDeleteAirdrop{}

func NewMsgDeleteAirdrop(distributor string, identifier string) *MsgDeleteAirdrop {
	return &MsgDeleteAirdrop{
		Distributor: distributor,
		Identifier:  identifier,
	}
}

func (msg *MsgDeleteAirdrop) Route() string {
	return RouterKey
}

func (msg *MsgDeleteAirdrop) Type() string {
	return TypeMsgDeleteAirdrop
}

func (msg *MsgDeleteAirdrop) GetSigners() []sdk.AccAddress {
	distributor, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{distributor}
}

func (msg *MsgDeleteAirdrop) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteAirdrop) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address (%s)", err)
	}

	if msg.Identifier == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "airdrop identifier not set")
	}
	return nil
}
