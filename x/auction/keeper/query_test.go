package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/auction/types"
)

func (s *KeeperTestSuite) TestQueryAuction() {
	// Create an auction
	expectedAuction := types.Auction{
		Type:                      types.AuctionType_AUCTION_TYPE_FCFS,
		Name:                      "test-auction",
		SellingDenom:              "ustrd",
		PaymentDenom:              "uatom",
		Enabled:                   true,
		MinPriceMultiplier:        sdkmath.LegacyNewDec(1),
		MinBidAmount:              sdkmath.NewInt(1000),
		Beneficiary:               "beneficiary-address",
		TotalSellingTokenSold:     sdkmath.ZeroInt(),
		TotalPaymentTokenReceived: sdkmath.ZeroInt(),
	}
	err := s.App.AuctionKeeper.SetAuction(s.Ctx, &expectedAuction)
	s.Require().NoError(err, "no error expected when setting auction")

	// Query for the auction
	req := &types.QueryAuctionRequest{
		Name: expectedAuction.Name,
	}
	resp, err := s.App.AuctionKeeper.Auction(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying auction")
	s.Require().Equal(expectedAuction, resp.Auction, "auction")

	// Query with invalid request
	_, err = s.App.AuctionKeeper.Auction(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")

	// Query with non-existent auction
	reqNonExistent := &types.QueryAuctionRequest{
		Name: "non-existent-auction",
	}
	_, err = s.App.AuctionKeeper.Auction(sdk.WrapSDKContext(s.Ctx), reqNonExistent)
	s.Require().Error(err, "error expected when querying non-existent auction")
	s.Require().Contains(err.Error(), "auction not found")
}

func (s *KeeperTestSuite) TestQueryAuctions() {
	// Create multiple auctions
	expectedAuctions := []types.Auction{
		{
			Type:                      types.AuctionType_AUCTION_TYPE_FCFS,
			Name:                      "test-auction-1",
			SellingDenom:              "ustrd",
			PaymentDenom:              "uatom",
			Enabled:                   true,
			MinPriceMultiplier:        sdkmath.LegacyNewDec(1),
			MinBidAmount:              sdkmath.NewInt(1000),
			Beneficiary:               "beneficiary-address-1",
			TotalSellingTokenSold:     sdkmath.ZeroInt(),
			TotalPaymentTokenReceived: sdkmath.ZeroInt(),
		},
		{
			Type:                      types.AuctionType_AUCTION_TYPE_FCFS,
			Name:                      "test-auction-2",
			SellingDenom:              "ustrd",
			PaymentDenom:              "uosmo",
			Enabled:                   true,
			MinPriceMultiplier:        sdkmath.LegacyNewDec(2),
			MinBidAmount:              sdkmath.NewInt(2000),
			Beneficiary:               "beneficiary-address-2",
			TotalSellingTokenSold:     sdkmath.ZeroInt(),
			TotalPaymentTokenReceived: sdkmath.ZeroInt(),
		},
	}

	for _, auction := range expectedAuctions {
		err := s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)
		s.Require().NoError(err, "no error expected when setting auction %+v", auction)
	}

	// Query all auctions
	req := &types.QueryAuctionsRequest{}
	resp, err := s.App.AuctionKeeper.Auctions(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying all auctions")
	s.Require().Equal(expectedAuctions, resp.Auctions, "auctions")

	// Query with invalid request
	_, err = s.App.AuctionKeeper.Auctions(sdk.WrapSDKContext(s.Ctx), nil)
	s.Require().Error(err, "error expected when querying with nil request")
}
