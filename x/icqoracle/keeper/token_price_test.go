package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/icqoracle/types"
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
