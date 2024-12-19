package types

import (
	"errors"
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func ValidateCreateAuctionParams(
	auctionName string,
	auctionType AuctionType,
	sellingDenom string,
	paymentDenom string,
	minPriceMultiplier math.LegacyDec,
	minBidAmount math.Int,
	beneficiary string,
) error {
	if auctionName == "" {
		return errors.New("auction-name must be specified")
	}
	if _, ok := AuctionType_name[int32(auctionType)]; !ok {
		return fmt.Errorf("auction-type %d is invalid", auctionType)
	}
	if sellingDenom == "" {
		return errors.New("selling-denom must be specified")
	}
	if paymentDenom == "" {
		return errors.New("payment-denom must be specified")
	}
	if !(minPriceMultiplier.GT(math.LegacyZeroDec()) && minPriceMultiplier.LTE(math.LegacyOneDec())) {
		return errors.New("min-price-multiplier must be > 0 and <= 1 (0 > minPriceMultiplier >= 1)")
	}
	if minBidAmount.LT(math.ZeroInt()) {
		return errors.New("min-bid-amount must be >= 0")
	}
	if _, err := sdk.AccAddressFromBech32(beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid beneficiary address (%s)", err)
	}

	return nil
}
