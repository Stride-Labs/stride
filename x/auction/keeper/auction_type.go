package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
)

// Define a type for bid handler functions
type AuctionBidHandler func(ctx sdk.Context, k Keeper, auction *types.Auction, bid *types.MsgPlaceBid) error

// Map of auction types to their handlers
var bidHandlers = map[types.AuctionType]AuctionBidHandler{
	types.AuctionType_AUCTION_TYPE_FCFS: fcfsBidHandler,
}

// fcfsBidHandler handles bids for First Come First Serve auctions
func fcfsBidHandler(ctx sdk.Context, k Keeper, auction *types.Auction, bid *types.MsgPlaceBid) error {
	// Get token amount being auctioned off
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	balance := k.bankKeeper.GetBalance(ctx, moduleAddr, auction.SellingDenom)
	tokenAmountAvailable := balance.Amount

	// Verify auction has enough tokens to service the bid amount
	if bid.PaymentTokenAmount.GT(tokenAmountAvailable) {
		return fmt.Errorf("bid wants %s%s but auction has only %s%s",
			bid.PaymentTokenAmount.String(),
			auction.SellingDenom,
			tokenAmountAvailable.String(),
			auction.SellingDenom,
		)
	}

	price, err := k.icqoracleKeeper.GetTokenPriceForQuoteDenom(ctx, auction.SellingDenom, auction.PaymentDenom)
	if err != nil {
		return err
	}

	discountedPrice := price.Mul(auction.PriceMultiplier)

	if bid.SellingTokenAmount.ToLegacyDec().
		Mul(discountedPrice).
		LT(bid.PaymentTokenAmount.ToLegacyDec()) {
		return fmt.Errorf("bid price too low: offered %s%s for %s%s, minimum required is %s%s (price=%s %s/%s)",
			bid.PaymentTokenAmount.String(),
			auction.PaymentDenom,
			bid.SellingTokenAmount.String(),
			auction.SellingDenom,
			bid.SellingTokenAmount.ToLegacyDec().Mul(discountedPrice).String(),
			auction.PaymentDenom,
			discountedPrice.String(),
			auction.PaymentDenom,
			auction.SellingDenom,
		)
	}

	// Safe to use MustAccAddressFromBech32 because bid.Bidder passed ValidateBasic
	bidder := sdk.MustAccAddressFromBech32(bid.Bidder)

	// Send paymentToken to beneficiary
	err = k.bankKeeper.SendCoins(
		ctx,
		bidder,
		sdk.MustAccAddressFromBech32(auction.Beneficiary),
		sdk.NewCoins(sdk.NewCoin(auction.PaymentDenom, bid.PaymentTokenAmount)),
	)
	if err != nil {
		return fmt.Errorf("failed to send payment tokens from bidder '%s' to beneficiary '%s': %w",
			bid.Bidder,
			auction.Beneficiary,
			err,
		)
	}

	// Send sellingToken to bidder
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		bidder,
		sdk.NewCoins(sdk.NewCoin(auction.SellingDenom, bid.SellingTokenAmount)),
	)
	if err != nil {
		return fmt.Errorf("failed to send auction tokens from module '%s' to bidder '%s': %w",
			types.ModuleName,
			bid.Bidder,
			err,
		)
	}

	auction.TotalSellingTokenSold = auction.TotalSellingTokenSold.Add(bid.SellingTokenAmount)
	auction.TotalPaymentTokenReceived = auction.TotalPaymentTokenReceived.Add(bid.PaymentTokenAmount)

	err = k.SetAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("failed to update auction stats")
	}

	// TODO emit event

	return nil
}
