package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/auction/types"
)

// Helper function to create 5 auction objects with various attributes
func (s *KeeperTestSuite) createAuctions() []types.Auction {
	auctions := []types.Auction{}
	for i := int64(1); i <= 5; i++ {
		auction := types.Auction{
			Type:                      types.AuctionType_AUCTION_TYPE_FCFS,
			Name:                      fmt.Sprintf("auction-%d", i),
			SellingDenom:              fmt.Sprintf("selling-%d", i),
			PaymentDenom:              fmt.Sprintf("payment-%d", i),
			Enabled:                   true,
			MinPriceMultiplier:        sdk.ZeroDec(),
			MinBidAmount:              sdkmath.NewInt(i),
			TotalPaymentTokenReceived: sdkmath.NewInt(i),
			TotalSellingTokenSold:     sdkmath.NewInt(i),
		}

		auctions = append(auctions, auction)
		s.App.AuctionKeeper.SetAuction(s.Ctx, &auction)
	}
	return auctions
}

// Tests Get/Set Auction
func (s *KeeperTestSuite) TestGetAuction() {
	auctions := s.createAuctions()

	for _, expected := range auctions {
		actual, err := s.App.AuctionKeeper.GetAuction(s.Ctx, expected.Name)
		s.Require().NoError(err, "auction %s should have been found", expected.Name)
		s.Require().Equal(expected, *actual, "auction %s", expected.Name)
	}

	_, err := s.App.AuctionKeeper.GetAuction(s.Ctx, "non-existent")
	s.Require().ErrorContains(err, "auction not found")
}

// Tests getting all auctions
func (s *KeeperTestSuite) TestGetAllAuctions() {
	expectedAuctions := s.createAuctions()

	actualAuctions := s.App.AuctionKeeper.GetAllAuctions(s.Ctx)
	s.Require().Equal(len(actualAuctions), len(expectedAuctions), "number of auctions")

	for i, expectedAuction := range expectedAuctions {
		s.Require().Equal(expectedAuction, actualAuctions[i], "auction %s", expectedAuction.Name)
	}
}
