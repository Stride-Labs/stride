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

func NewMsgPlaceBid(
	bidder string,
	tokenDenom string,
	auctionTokenAmount uint64,
	paymentTokenAmount uint64,
) *MsgPlaceBid {
	return &MsgPlaceBid{
		Bidder:             bidder,
		AuctionTokenDenom:  tokenDenom,
		AuctionTokenAmount: math.NewIntFromUint64(auctionTokenAmount),
		PaymentTokenAmount: math.NewIntFromUint64(paymentTokenAmount),
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
	if msg.AuctionTokenDenom == "" {
		return errors.New("token-denom must be specified")
	}
	if msg.AuctionTokenAmount.IsZero() {
		return errors.New("utoken-amount cannot be 0")
	}
	if msg.PaymentTokenAmount.IsZero() {
		return errors.New("ustrd-amount cannot be 0")
	}

	return nil
}

// ----------------------------------------------
//               MsgCreateAuction
// ----------------------------------------------

func NewMsgCreateAuction(
	admin string,
	auctionType AuctionType,
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
		AuctionType:     auctionType,
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
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid admin address (%s)", err)
	}
	if _, ok := AuctionType_name[int32(msg.AuctionType)]; !ok {
		return fmt.Errorf("auction-type %d is invalid", msg.AuctionType)
	}
	if msg.Denom == "" {
		return errors.New("denom must be specified")
	}
	if msg.MinBidAmount.LT(math.ZeroInt()) {
		return errors.New("min-bid-amount must be >= 0")
	}
	if !(msg.PriceMultiplier.GT(math.LegacyZeroDec()) && msg.PriceMultiplier.LTE(math.LegacyOneDec())) {
		return errors.New("price-multiplier must be > 0 and <= 1 (0 > priceMultiplier >= 1)")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid beneficiary address (%s)", err)
	}

	return nil
}

// ----------------------------------------------
//               MsgUpdateAuction
// ----------------------------------------------

func NewMsgUpdateAuction(
	admin string,
	auctionType AuctionType,
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
		AuctionType:     auctionType,
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
	if _, ok := AuctionType_name[int32(msg.AuctionType)]; !ok {
		return errors.New("auction-type is invalid")
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
