package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v29/utils"
	"github.com/Stride-Labs/stride/v29/x/auction/types"
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
	sellingAmountAvailable := balance.Amount

	// Verify auction has enough selling tokens to service the bid
	if bid.SellingTokenAmount.GT(sellingAmountAvailable) {
		return fmt.Errorf("bid wants to buy %s%s but auction only has %s%s",
			bid.SellingTokenAmount.String(),
			auction.SellingDenom,
			sellingAmountAvailable.String(),
			auction.SellingDenom,
		)
	}

	// Note: price converts SellingToken to PaymentToken
	// Any calculation down the road makes sense only if price is multiplied by a derivative of SellingToken
	price, err := k.icqoracleKeeper.GetTokenPriceForQuoteDenom(ctx, auction.SellingDenom, auction.PaymentDenom)
	if err != nil {
		return errorsmod.Wrapf(err, "error getting price for baseDenom='%s' quoteDenom='%s'", auction.SellingDenom, auction.PaymentDenom)
	}

	// Apply MinPriceMultiplier
	bidsFloorPrice := price.Mul(auction.MinPriceMultiplier)
	minPaymentRequired := bid.SellingTokenAmount.ToLegacyDec().Mul(bidsFloorPrice)

	// if paymentAmount < sellingAmount * bidsFloorPrice
	if bid.PaymentTokenAmount.ToLegacyDec().LT(minPaymentRequired) {
		return fmt.Errorf("bid price too low: offered %s%s for %s%s, bids floor price is %s%s (price=%s %s/%s)",
			bid.PaymentTokenAmount.String(),
			auction.PaymentDenom,
			bid.SellingTokenAmount.String(),
			auction.SellingDenom,
			minPaymentRequired.String(),
			auction.PaymentDenom,
			bidsFloorPrice.String(),
			auction.PaymentDenom,
			auction.SellingDenom,
		)
	}

	// Safe to use MustAccAddressFromBech32 because bid.Bidder passed ValidateBasic
	bidder := sdk.MustAccAddressFromBech32(bid.Bidder)

	// Send paymentToken to beneficiary
	// Note: checkBlockedAddr=false because beneficiary can be a module
	err = utils.SafeSendCoins(
		false,
		k.bankKeeper,
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

	k.SetAuction(ctx, auction)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBidAccepted,
			sdk.NewAttribute(types.AttributeKeyAuctionName, auction.Name),
			sdk.NewAttribute(types.AttributeKeyBidder, bid.Bidder),
			sdk.NewAttribute(types.AttributeKeyPaymentAmount, bid.PaymentTokenAmount.String()),
			sdk.NewAttribute(types.AttributeKeyPaymentDenom, auction.PaymentDenom),
			sdk.NewAttribute(types.AttributeKeySellingAmount, bid.SellingTokenAmount.String()),
			sdk.NewAttribute(types.AttributeKeySellingDenom, auction.SellingDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, bidsFloorPrice.String()),
		),
	)

	return nil
}
