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
	priceMultiplier math.LegacyDec,
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
	if !(priceMultiplier.GT(math.LegacyZeroDec()) && priceMultiplier.LTE(math.LegacyOneDec())) {
		return errors.New("price-multiplier must be > 0 and <= 1 (0 > priceMultiplier >= 1)")
	}
	if minBidAmount.LT(math.ZeroInt()) {
		return errors.New("min-bid-amount must be >= 0")
	}
	if _, err := sdk.AccAddressFromBech32(beneficiary); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid beneficiary address (%s)", err)
	}

	return nil
}
