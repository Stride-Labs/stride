package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

// Helper function to create 5 tokenPrice objects with various attributes
func (s *KeeperTestSuite) createTokenPrices() []types.TokenPrice {
	tokenPrices := []types.TokenPrice{}
	for i := int64(1); i <= 5; i++ {
		tokenPrice := types.TokenPrice{
			BaseDenom:  fmt.Sprintf("base-%d", i),
			QuoteDenom: fmt.Sprintf("quote-%d", i),
			SpotPrice:  sdk.ZeroDec(),
		}

		tokenPrices = append(tokenPrices, tokenPrice)
		s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	}
	return tokenPrices
}

// Tests Get/Set TokenPrice
func (s *KeeperTestSuite) TestGetTokenPrice() {
	tokenPrices := s.createTokenPrices()

	for _, expected := range tokenPrices {
		actual, err := s.App.ICQOracleKeeper.GetTokenPrice(s.Ctx, expected.BaseDenom, expected.QuoteDenom, expected.OsmosisPoolId)
		s.Require().NoError(err, "tokenPrice %s should have been found", expected.BaseDenom)
		s.Require().Equal(expected, actual)
	}

	_, err := s.App.ICQOracleKeeper.GetTokenPrice(s.Ctx, "non-existent", "non-existent", 0)
	s.Require().ErrorContains(err, "price not found")
}

// Tests getting all tokenPrices
func (s *KeeperTestSuite) TestGetAllTokenPrices() {
	expectedTokenPrices := s.createTokenPrices()

	actualTokenPrices := s.App.ICQOracleKeeper.GetAllTokenPrices(s.Ctx)
	s.Require().Equal(len(actualTokenPrices), len(expectedTokenPrices), "number of tokenPrices")

	for i, expectedTokenPrice := range expectedTokenPrices {
		s.Require().Equal(expectedTokenPrice, actualTokenPrices[i])
	}
}

// Tests getting a price from a common quote denom
func (s *KeeperTestSuite) TestGetTokenPriceForQuoteDenom() {
	freshTime := s.Ctx.BlockTime().Add(-1 * time.Second)
	// staleTime := s.Ctx.BlockTime().Add(-1 * time.Hour)

	testCases := []struct {
		name          string
		baseDenom     string
		quoteDenom    string
		tokenPrices   []types.TokenPrice
		expectedPrice sdk.Dec
		expectedError string
	}{
		{
			name:       "exact price found",
			baseDenom:  "denomA",
			quoteDenom: "denomB",
			tokenPrices: []types.TokenPrice{
				{BaseDenom: "denomA", QuoteDenom: "denomB", SpotPrice: sdk.NewDec(4), LastRequestTime: freshTime},
			},
			expectedPrice: sdk.MustNewDecFromStr("4.0"),
		},
		{
			name:       "exact price found with inversion",
			baseDenom:  "denomA",
			quoteDenom: "denomB",
			tokenPrices: []types.TokenPrice{
				{BaseDenom: "denomB", QuoteDenom: "denomA", SpotPrice: sdk.NewDec(4), LastRequestTime: freshTime},
			},
			expectedPrice: sdk.MustNewDecFromStr("0.25"), // 1 / price = 1 / 4
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			params := types.DefaultParams()
			params.PriceExpirationTimeoutSec = 10 * 60 // 10 minutes
			s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

			for _, tokenPrice := range tc.tokenPrices {
				tokenPrice.OsmosisPoolId = 1
				s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
			}

			actualPrice, actualError := s.App.ICQOracleKeeper.GetTokenPriceForQuoteDenom(s.Ctx, tc.baseDenom, tc.quoteDenom)
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError)
			} else {
				s.Require().NoError(actualError)
				s.Require().Equal(tc.expectedPrice, actualPrice, "price")
			}
		})
	}
}
