package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	icqoracleKeeper types.IcqOracleKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	icqoracleKeeper types.IcqOracleKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		icqoracleKeeper: icqoracleKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetAuction stores auction info for a token
func (k Keeper) SetAuction(ctx sdk.Context, auction types.Auction) error {
	store := ctx.KVStore(k.storeKey)
	key := types.AuctionKey(auction.Denom)
	bz, err := k.cdc.Marshal(&auction)
	if err != nil {
		return fmt.Errorf("error setting auction for denom '%s': %w", auction.Denom, err)
	}

	store.Set(key, bz)
	return nil
}

// GetAuction retrieves auction info for a token
func (k Keeper) GetAuction(ctx sdk.Context, denom string) (types.Auction, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.AuctionKey(denom)

	bz := store.Get(key)
	if bz == nil {
		return types.Auction{}, fmt.Errorf("auction not found for denom '%s'", denom)
	}

	var auction types.Auction
	if err := k.cdc.Unmarshal(bz, &auction); err != nil {
		return types.Auction{}, fmt.Errorf("error retrieving auction for denom '%s': %w", auction.Denom, err)
	}

	return auction, nil
}

// GetAllAuctions retrieves all stored auctions
func (k Keeper) GetAllAuctions(ctx sdk.Context) []types.Auction {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.KeyAuctionPrefix))
	defer iterator.Close()

	auctions := []types.Auction{}
	for ; iterator.Valid(); iterator.Next() {
		var auction types.Auction
		k.cdc.MustUnmarshal(iterator.Value(), &auction)
		auctions = append(auctions, auction)
	}

	return auctions
}

// PlaceBid places an auction bid and executes it based on the auction type
func (k Keeper) PlaceBid(ctx sdk.Context, bid *types.MsgPlaceBid) error {
	// Get auction
	auction, err := k.GetAuction(ctx, bid.AuctionTokenDenom)
	if err != nil {
		return fmt.Errorf("cannot get auction for denom='%s': %w", bid.AuctionTokenDenom, err)
	}

	// TODO check auction type and call handler

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

	auctionTokenPrices, err := k.icqoracleKeeper.GetTokenPricesByDenom(ctx, auction.Denom)
	if err != nil {
		return fmt.Errorf("error getting price for '%s': %w", auction.Denom, err)
	}
	if len(auctionTokenPrices) == 0 {
		return fmt.Errorf("no price for '%s'", auction.Denom)
	}

	paymentTokenDenom := "ustrd" // TODO fix
	paymentTokenPrices, err := k.icqoracleKeeper.GetTokenPricesByDenom(ctx, paymentTokenDenom)
	if err != nil {
		return fmt.Errorf("error getting price for '%s': %w", paymentTokenDenom, err)
	}
	if len(paymentTokenPrices) == 0 {
		return fmt.Errorf("no price for '%s'", paymentTokenDenom)
	}

	// Get price staleness timeout
	priceStaleTimeoutSec := int64(k.icqoracleKeeper.GetParams(ctx).PriceStaleTimeoutSec)

	// Find a common quote denom and calculate auctionToken-paymentToken price
	auctionTokenToPaymentTokenPrice := math.LegacyZeroDec()
	foundCommonQuoteToken := false
	foundAuctionTokenStalePrice := false
	foundPaymentTokenStalePrice := false
	for quoteDenom1, auctionTokenPrice := range auctionTokenPrices {
		for quoteDenom2, paymentTokenPrice := range paymentTokenPrices {
			if quoteDenom1 == quoteDenom2 {
				foundCommonQuoteToken = true

				// Check that prices are not stale
				if ctx.BlockTime().Unix()-auctionTokenPrice.UpdatedAt.Unix() > priceStaleTimeoutSec {
					foundAuctionTokenStalePrice = true
					continue
				}
				if ctx.BlockTime().Unix()-paymentTokenPrice.UpdatedAt.Unix() > priceStaleTimeoutSec {
					foundPaymentTokenStalePrice = true
					continue
				}

				// Calculate the price of 1 auctionToken in paymentToken including auction discount
				// i.e. baseToken = auctionToken, quoteToken = paymentToken
				auctionTokenToPaymentTokenPrice = auctionTokenPrice.SpotPrice.Quo(paymentTokenPrice.SpotPrice).Mul(auction.PriceMultiplier)
				break
			}
		}
	}

	if auctionTokenToPaymentTokenPrice.IsZero() {
		return fmt.Errorf(
			"could not calculate price for baseToken='%s' quoteToken='%s' (foundCommonQuoteToken='%v', foundAuctionTokenStalePrice='%v', foundPaymentTokenStalePrice='%v')",
			auction.Denom,
			paymentTokenDenom,
			foundCommonQuoteToken,
			foundAuctionTokenStalePrice,
			foundPaymentTokenStalePrice,
		)
	}

	if bid.AuctionTokenAmount.ToLegacyDec().Mul(auctionTokenToPaymentTokenPrice).LT(bid.PaymentTokenAmount.ToLegacyDec()) {
		return fmt.Errorf("bid price too low: offered %s%s for %s%s, minimum required is %s%s (price=%s %s/%s)",
			bid.PaymentTokenAmount.String(),
			paymentTokenDenom,
			bid.AuctionTokenAmount.String(),
			auction.Denom,
			bid.AuctionTokenAmount.ToLegacyDec().Mul(auctionTokenToPaymentTokenPrice).String(),
			paymentTokenDenom,
			auctionTokenToPaymentTokenPrice.String(),
			paymentTokenDenom,
			auction.Denom,
		)
	}

	// Execute the exchange

	// Safe to use MustAccAddressFromBech32 because bid.Bidder passed ValidateBasic
	bidder := sdk.MustAccAddressFromBech32(bid.Bidder)

	// Send paymentToken to beneficiary
	// TODO move this up to fail early if not enought funds?
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
	// TODO move this up to fail early if not enought funds?
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
