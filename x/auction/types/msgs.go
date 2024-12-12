package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgPlaceBid      = "place_bid"
	TypeMsgCreateAuction = "create_auction"
	TypeMsgUpdateAuction = "update_auction"
)

var (
	_ sdk.Msg = &MsgPlaceBid{}
	_ sdk.Msg = &MsgCreateAuction{}
	_ sdk.Msg = &MsgUpdateAuction{}

	// Implement legacy interface for ledger support
	_ legacytx.LegacyMsg = &MsgPlaceBid{}
	_ legacytx.LegacyMsg = &MsgCreateAuction{}
	_ legacytx.LegacyMsg = &MsgUpdateAuction{}
)

// ----------------------------------------------
//               MsgPlaceBid
// ----------------------------------------------

func NewMsgPlaceBid(bidder, tokenDenom string, utokenAmount, ustrdAmount uint64) *MsgPlaceBid {
	return &MsgPlaceBid{
		Bidder:       bidder,
		TokenDenom:   tokenDenom,
		UtokenAmount: math.NewIntFromUint64(utokenAmount),
		UstrdAmount:  math.NewIntFromUint64(ustrdAmount),
	}
}

func (msg MsgPlaceBid) Type() string {
	return TypeMsgPlaceBid
}

func (msg MsgPlaceBid) Route() string {
	return RouterKey
}

func (msg *MsgPlaceBid) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Bidder)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgPlaceBid) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPlaceBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Bidder); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.TokenDenom == "" {
		return errors.New("token-denom must be specified")
	}
	if msg.UtokenAmount.IsZero() {
		return errors.New("utoken-amount cannot be 0")
	}
	if msg.UstrdAmount.IsZero() {
		return errors.New("ustrd-amount cannot be 0")
	}

	return nil
}

// ----------------------------------------------
//               MsgCreateAuction
// ----------------------------------------------

func NewMsgCreateAuction(admin string,
	denom string,
	enabled bool,
	priceMultiplier string,
	minBidAmount uint64,
	beneficiary string,
) *MsgCreateAuction {
	priceMultiplierDec, err := math.LegacyNewDecFromStr(priceMultiplier)
	if err != nil {
		panic(fmt.Sprintf("cannot parse LegacyDecimal from priceMultiplier '%s'", priceMultiplier))
	}

	return &MsgCreateAuction{
		Admin:           admin,
		Denom:           denom,
		Enabled:         enabled,
		PriceMultiplier: priceMultiplierDec,
		MinBidAmount:    math.NewIntFromUint64(minBidAmount),
		Beneficiary:     beneficiary,
	}
}

func (msg MsgCreateAuction) Type() string {
	return TypeMsgCreateAuction
}

func (msg MsgCreateAuction) Route() string {
	return RouterKey
}

func (msg *MsgCreateAuction) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgCreateAuction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateAuction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.Denom == "" {
		return errors.New("denom must be specified")
	}
	if msg.MinBidAmount.LT(math.ZeroInt()) {
		return errors.New("min-bid-amount must be at least 0")
	}
	if msg.PriceMultiplier.IsZero() {
		return errors.New("price-multiplier cannot be 0")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgUpdateAuction
// ----------------------------------------------

func NewMsgUpdateAuction(admin string,
	denom string,
	enabled bool,
	priceMultiplier string,
	minBidAmount uint64,
	beneficiary string,
) *MsgUpdateAuction {
	priceMultiplierDec, err := math.LegacyNewDecFromStr(priceMultiplier)
	if err != nil {
		panic(fmt.Sprintf("cannot parse LegacyDecimal from priceMultiplier '%s'", priceMultiplier))
	}

	return &MsgUpdateAuction{
		Admin:           admin,
		Denom:           denom,
		Enabled:         enabled,
		PriceMultiplier: priceMultiplierDec,
		MinBidAmount:    math.NewIntFromUint64(minBidAmount),
		Beneficiary:     beneficiary,
	}
}

func (msg MsgUpdateAuction) Type() string {
	return TypeMsgUpdateAuction
}

func (msg MsgUpdateAuction) Route() string {
	return RouterKey
}

func (msg *MsgUpdateAuction) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

func (msg *MsgUpdateAuction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateAuction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.Denom == "" {
		return errors.New("denom must be specified")
	}
	if msg.MinBidAmount.LT(math.ZeroInt()) {
		return errors.New("min-bid-amount must be at least 0")
	}
	if msg.PriceMultiplier.IsZero() {
		return errors.New("price-multiplier cannot be 0")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}
