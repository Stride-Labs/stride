package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v30/utils"
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
)

// ----------------------------------------------
//               MsgPlaceBid
// ----------------------------------------------

func NewMsgPlaceBid(
	bidder string,
	AuctionName string,
	sellingTokenAmount sdkmath.Int,
	paymentTokenAmount sdkmath.Int,
) *MsgPlaceBid {
	return &MsgPlaceBid{
		Bidder:             bidder,
		AuctionName:        AuctionName,
		SellingTokenAmount: sellingTokenAmount,
		PaymentTokenAmount: paymentTokenAmount,
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

func (msg *MsgPlaceBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Bidder); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.AuctionName == "" {
		return errors.New("auction-name must be specified")
	}
	if msg.SellingTokenAmount.IsZero() {
		return errors.New("selling-token-amount cannot be 0")
	}
	if msg.PaymentTokenAmount.IsZero() {
		return errors.New("payment-token-amount cannot be 0")
	}

	return nil
}

// ----------------------------------------------
//               MsgCreateAuction
// ----------------------------------------------

func NewMsgCreateAuction(
	admin string,
	auctionName string,
	auctionType AuctionType,
	sellingDenom string,
	paymentDenom string,
	enabled bool,
	minPriceMultiplier string,
	minBidAmount uint64,
	beneficiary string,
) *MsgCreateAuction {
	minPriceMultiplierDec, err := sdkmath.LegacyNewDecFromStr(minPriceMultiplier)
	if err != nil {
		panic(fmt.Sprintf("cannot parse LegacyDecimal from minPriceMultiplier '%s'", minPriceMultiplier))
	}

	return &MsgCreateAuction{
		Admin:              admin,
		AuctionName:        auctionName,
		AuctionType:        auctionType,
		SellingDenom:       sellingDenom,
		PaymentDenom:       paymentDenom,
		Enabled:            enabled,
		MinPriceMultiplier: minPriceMultiplierDec,
		MinBidAmount:       sdkmath.NewIntFromUint64(minBidAmount),
		Beneficiary:        beneficiary,
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

func (msg *MsgCreateAuction) ValidateBasic() error {
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}

	return ValidateCreateAuctionParams(
		msg.AuctionName,
		msg.AuctionType,
		msg.SellingDenom,
		msg.PaymentDenom,
		msg.MinPriceMultiplier,
		msg.MinBidAmount,
		msg.Beneficiary,
	)
}

// ----------------------------------------------
//               MsgUpdateAuction
// ----------------------------------------------

func NewMsgUpdateAuction(
	admin string,
	auctionName string,
	auctionType AuctionType,
	enabled bool,
	minPriceMultiplier string,
	minBidAmount uint64,
	beneficiary string,
) *MsgUpdateAuction {
	minPriceMultiplierDec, err := sdkmath.LegacyNewDecFromStr(minPriceMultiplier)
	if err != nil {
		panic(fmt.Sprintf("cannot parse LegacyDecimal from minPriceMultiplier '%s'", minPriceMultiplier))
	}

	return &MsgUpdateAuction{
		Admin:              admin,
		AuctionName:        auctionName,
		AuctionType:        auctionType,
		Enabled:            enabled,
		MinPriceMultiplier: minPriceMultiplierDec,
		MinBidAmount:       sdkmath.NewIntFromUint64(minBidAmount),
		Beneficiary:        beneficiary,
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

func (msg *MsgUpdateAuction) ValidateBasic() error {
	if err := utils.ValidateAdminAddress(msg.Admin); err != nil {
		return err
	}
	if msg.AuctionName == "" {
		return errors.New("auction-name must be specified")
	}
	if _, ok := AuctionType_name[int32(msg.AuctionType)]; !ok {
		return errors.New("auction-type is invalid")
	}
	if msg.MinPriceMultiplier.IsZero() {
		return errors.New("min-price-multiplier cannot be 0")
	}
	if msg.MinBidAmount.LT(sdkmath.ZeroInt()) {
		return errors.New("min-bid-amount must be at least 0")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}

	return nil
}
