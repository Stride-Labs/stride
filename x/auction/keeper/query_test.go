package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Stride-Labs/stride/v26/x/auction/types"
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
	s.App.AuctionKeeper.SetAuction(s.Ctx, &expectedAuction)

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
		s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)
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

func (s *KeeperTestSuite) TestQueryAuctionsPagination() {
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
		{
			Type:                      types.AuctionType_AUCTION_TYPE_FCFS,
			Name:                      "test-auction-3",
			SellingDenom:              "ustrd",
			PaymentDenom:              "ujuno",
			Enabled:                   true,
			MinPriceMultiplier:        sdkmath.LegacyNewDec(3),
			MinBidAmount:              sdkmath.NewInt(3000),
			Beneficiary:               "beneficiary-address-3",
			TotalSellingTokenSold:     sdkmath.ZeroInt(),
			TotalPaymentTokenReceived: sdkmath.ZeroInt(),
		},
	}

	for _, auction := range expectedAuctions {
		s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)
	}

	// Test pagination with limit of 2
	req := &types.QueryAuctionsRequest{
		Pagination: &query.PageRequest{
			Limit: 2,
		},
	}
	resp, err := s.App.AuctionKeeper.Auctions(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying with pagination")
	s.Require().Len(resp.Auctions, 2, "should return 2 auctions")
	s.Require().Equal(expectedAuctions[:2], resp.Auctions, "first page auctions")
	s.Require().NotNil(resp.Pagination.NextKey, "next key should be present")

	// Query second page
	req = &types.QueryAuctionsRequest{
		Pagination: &query.PageRequest{
			Key: resp.Pagination.NextKey,
		},
	}
	resp, err = s.App.AuctionKeeper.Auctions(sdk.WrapSDKContext(s.Ctx), req)
	s.Require().NoError(err, "no error expected when querying second page")
	s.Require().Len(resp.Auctions, 1, "should return 1 auction")
	s.Require().Equal(expectedAuctions[2:], resp.Auctions, "second page auctions")
	s.Require().Nil(resp.Pagination.NextKey, "next key should be nil")
}
