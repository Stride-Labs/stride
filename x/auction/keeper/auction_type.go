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
	balance := k.bankKeeper.GetBalance(ctx, moduleAddr, auction.Denom)
	tokenAmountAvailable := balance.Amount

	// Verify auction has enough tokens to service the bid amount
	if bid.AuctionTokenAmount.GT(tokenAmountAvailable) {
		return fmt.Errorf("bid wants %s%s but auction has only %s%s",
			bid.AuctionTokenAmount.String(),
			auction.Denom,
			tokenAmountAvailable.String(),
			auction.Denom,
		)
	}

	paymentTokenDenom := "ustrd" // TODO fix

	price, err := k.icqoracleKeeper.GetTokenPriceForQuoteDenom(ctx, auction.Denom, paymentTokenDenom)
	if err != nil {
		return err
	}

	discountedPrice := price.Mul(auction.PriceMultiplier)

	if bid.AuctionTokenAmount.ToLegacyDec().
		Mul(discountedPrice).
		LT(bid.PaymentTokenAmount.ToLegacyDec()) {
		return fmt.Errorf("bid price too low: offered %s%s for %s%s, minimum required is %s%s (price=%s %s/%s)",
			bid.PaymentTokenAmount.String(),
			paymentTokenDenom,
			bid.AuctionTokenAmount.String(),
			auction.Denom,
			bid.AuctionTokenAmount.ToLegacyDec().Mul(discountedPrice).String(),
			paymentTokenDenom,
			discountedPrice.String(),
			paymentTokenDenom,
			auction.Denom,
		)
	}

	// Safe to use MustAccAddressFromBech32 because bid.Bidder passed ValidateBasic
	bidder := sdk.MustAccAddressFromBech32(bid.Bidder)

	// Send paymentToken to beneficiary
	err = k.bankKeeper.SendCoins(
		ctx,
		bidder,
		sdk.MustAccAddressFromBech32(auction.Beneficiary),
		sdk.NewCoins(sdk.NewCoin(paymentTokenDenom, bid.PaymentTokenAmount)),
	)
	if err != nil {
		return fmt.Errorf("failed to send payment tokens from bidder '%s' to beneficiary '%s': %w",
			bid.Bidder,
			auction.Beneficiary,
			err,
		)
	}

	// Send auctionToken to bidder
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		bidder,
		sdk.NewCoins(sdk.NewCoin(auction.Denom, bid.AuctionTokenAmount)),
	)
	if err != nil {
		return fmt.Errorf("failed to send auction tokens from module '%s' to bidder '%s': %w",
			types.ModuleName,
			bid.Bidder,
			err,
		)
	}

	auction.TotalTokensSold = auction.TotalTokensSold.Add(bid.AuctionTokenAmount)
	auction.TotalStrdBurned = auction.TotalStrdBurned.Add(bid.PaymentTokenAmount)

	err = k.SetAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("failed to update auction stats")
	}

	// TODO emit event

	return nil
}
