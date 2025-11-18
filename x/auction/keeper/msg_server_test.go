package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v30/x/auction/types"
	icqoracletypes "github.com/Stride-Labs/stride/v30/x/icqoracle/types"
)

func (s *KeeperTestSuite) TestCreateAuction() {
	// Create a new auction
	msg := types.MsgCreateAuction{
		AuctionName:        "test-auction",
		AuctionType:        types.AuctionType_AUCTION_TYPE_FCFS,
		SellingDenom:       "ustrd",
		PaymentDenom:       "uatom",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyMustNewDecFromStr("0.95"),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        "beneficiary-address",
	}
	_, err := s.GetMsgServer().CreateAuction(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when creating auction")

	// Confirm the auction was created
	auction := s.MustGetAuction(msg.AuctionName)

	s.Require().Equal(msg.AuctionType, auction.Type, "auction type")
	s.Require().Equal(msg.AuctionName, auction.Name, "auction name")
	s.Require().Equal(msg.SellingDenom, auction.SellingDenom, "selling denom")
	s.Require().Equal(msg.PaymentDenom, auction.PaymentDenom, "payment denom")
	s.Require().Equal(msg.Enabled, auction.Enabled, "enabled")
	s.Require().Equal(msg.MinPriceMultiplier, auction.MinPriceMultiplier, "min price multiplier")
	s.Require().Equal(msg.MinBidAmount, auction.MinBidAmount, "min bid amount")
	s.Require().Equal(msg.Beneficiary, auction.Beneficiary, "beneficiary")
	s.Require().Equal(sdkmath.ZeroInt(), auction.TotalPaymentTokenReceived, "total payment token received")
	s.Require().Equal(sdkmath.ZeroInt(), auction.TotalSellingTokenSold, "total selling token sold")

	// Attempt to create it again, it should fail
	_, err = s.GetMsgServer().CreateAuction(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAuctionAlreadyExists)
}

func (s *KeeperTestSuite) TestUpdateAuction() {
	// Create an auction first
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "ustrd",
		PaymentDenom:       "uatom",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        "beneficiary-address",
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Update the auction
	msg := types.MsgUpdateAuction{
		AuctionName:        auction.Name,
		AuctionType:        types.AuctionType_AUCTION_TYPE_FCFS,
		Enabled:            false,
		MinPriceMultiplier: sdkmath.LegacyNewDec(2),
		MinBidAmount:       sdkmath.NewInt(2000),
		Beneficiary:        "new-beneficiary-address",
	}
	_, err := s.GetMsgServer().UpdateAuction(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when updating auction")

	// Confirm the auction was updated
	updatedAuction := s.MustGetAuction(msg.AuctionName)
	s.Require().Equal(msg.AuctionType, updatedAuction.Type, "auction type")
	s.Require().Equal(msg.Enabled, updatedAuction.Enabled, "enabled")
	s.Require().Equal(msg.MinPriceMultiplier, updatedAuction.MinPriceMultiplier, "min price multiplier")
	s.Require().Equal(msg.MinBidAmount, updatedAuction.MinBidAmount, "min bid amount")
	s.Require().Equal(msg.Beneficiary, updatedAuction.Beneficiary, "beneficiary")

	// Try to update non-existent auction, it should fail
	msg.AuctionName = "non-existent-auction"
	_, err = s.GetMsgServer().UpdateAuction(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorIs(err, types.ErrAuctionDoesntExist)
}

func (s *KeeperTestSuite) TestFcfsPlaceBidHappyPath() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Create a price
	tokenPrice := icqoracletypes.TokenPrice{
		BaseDenom:        auction.SellingDenom,
		QuoteDenom:       auction.PaymentDenom,
		OsmosisPoolId:    1,
		SpotPrice:        sdkmath.LegacyNewDec(1),
		LastResponseTime: s.Ctx.BlockTime(),
		QueryInProgress:  false,
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// Mint enough payment coins to bidder to pay
	s.FundAccount(bidder, sdk.NewCoin(auction.PaymentDenom, msg.PaymentTokenAmount))

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when placing bid")

	// Check payment token to beneficiary
	s.Require().Contains(s.Ctx.EventManager().Events(),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, msg.Bidder),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoins(sdk.NewCoin(auction.PaymentDenom, msg.PaymentTokenAmount)).String()),
		),
	)

	// Check selling token to bidder
	s.Require().Contains(s.Ctx.EventManager().Events(),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, msg.Bidder),
			sdk.NewAttribute(banktypes.AttributeKeySender, s.App.AccountKeeper.GetModuleAddress(types.ModuleName).String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, sdk.NewCoins(sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount)).String()),
		),
	)

	// Check PlaceBid events
	s.Require().Contains(s.Ctx.EventManager().Events(),
		sdk.NewEvent(
			types.EventTypeBidAccepted,
			sdk.NewAttribute(types.AttributeKeyAuctionName, auction.Name),
			sdk.NewAttribute(types.AttributeKeyBidder, msg.Bidder),
			sdk.NewAttribute(types.AttributeKeyPaymentAmount, msg.PaymentTokenAmount.String()),
			sdk.NewAttribute(types.AttributeKeyPaymentDenom, auction.PaymentDenom),
			sdk.NewAttribute(types.AttributeKeySellingAmount, msg.SellingTokenAmount.String()),
			sdk.NewAttribute(types.AttributeKeySellingDenom, auction.SellingDenom),
			sdk.NewAttribute(types.AttributeKeyPrice, sdkmath.LegacyNewDec(1).String()),
		),
	)
}

func (s *KeeperTestSuite) TestFcfsPlaceBidUnsupportedAuctionType() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_UNSPECIFIED,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unsupported auction type")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidAuctionNoFound() {
	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        "banana",
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "cannot get auction for name='banana'")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidNotEnoughSellingTokens() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "bid wants to buy 1000uosmo but auction only has 0uosmo")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidNoPriceForSellingDenom() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "error getting price for baseDenom='uosmo' quoteDenom='ustrd': no price found for baseDenom 'uosmo'")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidNoPriceForPaymentDenom() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Create a price only for SellingDenom
	tokenPrice := icqoracletypes.TokenPrice{
		BaseDenom:        auction.SellingDenom,
		QuoteDenom:       "uusdc",
		OsmosisPoolId:    1,
		SpotPrice:        sdkmath.LegacyNewDec(1),
		LastResponseTime: s.Ctx.BlockTime(),
		QueryInProgress:  false,
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "error getting price for baseDenom='uosmo' quoteDenom='ustrd': no price found for baseDenom 'uosmo'")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidTooLowPrice() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        "beneficiary-address",
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Create a price
	tokenPrice := icqoracletypes.TokenPrice{
		BaseDenom:        auction.SellingDenom,
		QuoteDenom:       auction.PaymentDenom,
		OsmosisPoolId:    1,
		SpotPrice:        sdkmath.LegacyNewDec(1),
		LastResponseTime: s.Ctx.BlockTime(),
		QueryInProgress:  false,
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Prepare bid with a price that's too low
	// With spot price = 1 and multiplier = 1, minimum accepted price should be 1
	// Setting bid with price < 1
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName: auction.Name,
		Bidder:      bidder.String(),
		// Make the effective price 0.5, which is below floor of 1
		SellingTokenAmount: sdkmath.NewInt(2000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// Mint enough payment coins to bidder to pay
	s.FundAccount(bidder, sdk.NewCoin(auction.PaymentDenom, msg.PaymentTokenAmount))

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "bid price too low")
}

func (s *KeeperTestSuite) TestFcfsPlaceBidNotEnoughPaymentTokens() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Create a price
	tokenPrice := icqoracletypes.TokenPrice{
		BaseDenom:        auction.SellingDenom,
		QuoteDenom:       auction.PaymentDenom,
		OsmosisPoolId:    1,
		SpotPrice:        sdkmath.LegacyNewDec(1),
		LastResponseTime: s.Ctx.BlockTime(),
		QueryInProgress:  false,
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// DON'T fund the bidder with payment tokens

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "failed to send payment tokens from bidder")
}

func (s *KeeperTestSuite) TestPlaceBidLessThanMinBidAmount() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Prepare bid with amount lower than min bid amount
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(500), // less than min bid amount
	}

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "payment bid amount '500' is less than the minimum bid '1000' amount for auction 'test-auction'")
}

func (s *KeeperTestSuite) TestPlaceBidExactMinBidAmount() {
	// Create an auction
	auction := types.Auction{
		Type:               types.AuctionType_AUCTION_TYPE_FCFS,
		Name:               "test-auction",
		SellingDenom:       "uosmo",
		PaymentDenom:       "ustrd",
		Enabled:            true,
		MinPriceMultiplier: sdkmath.LegacyNewDec(1),
		MinBidAmount:       sdkmath.NewInt(1000),
		Beneficiary:        s.App.StrdBurnerKeeper.GetStrdBurnerAddress().String(),
	}
	s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)

	// Create a price
	tokenPrice := icqoracletypes.TokenPrice{
		BaseDenom:        auction.SellingDenom,
		QuoteDenom:       auction.PaymentDenom,
		OsmosisPoolId:    1,
		SpotPrice:        sdkmath.LegacyNewDec(1),
		LastResponseTime: s.Ctx.BlockTime(),
		QueryInProgress:  false,
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Prepare bid
	bidder := s.TestAccs[0]
	msg := types.MsgPlaceBid{
		AuctionName:        auction.Name,
		Bidder:             bidder.String(),
		SellingTokenAmount: sdkmath.NewInt(1000),
		PaymentTokenAmount: sdkmath.NewInt(1000),
	}

	// Mint enough selling coins to auction module to sell
	s.FundModuleAccount(types.ModuleName, sdk.NewCoin(auction.SellingDenom, msg.SellingTokenAmount))

	// Mint enough payment coins to bidder to pay
	s.FundAccount(bidder, sdk.NewCoin(auction.PaymentDenom, msg.PaymentTokenAmount))

	// Place Bid
	_, err := s.GetMsgServer().PlaceBid(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when placing bid")
}
