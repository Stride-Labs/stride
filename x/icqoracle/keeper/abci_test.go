package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v27/x/interchainquery/types"
)

func (s *KeeperTestSuite) TestBeginBlockerSubmitICQ() {
	var submitICQCalled bool

	// Mock ICQ keeper to capture submitted queries
	mockICQKeeper := MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			submitICQCalled = true
			return nil
		},
	}

	params := types.Params{
		UpdateIntervalSec: 60, // 1 minute interval
	}

	now := time.Now().UTC()
	staleTime := now.Add(-2 * time.Minute)  // Older than update interval
	freshTime := now.Add(-30 * time.Second) // More recent than update interval

	testCases := []struct {
		name              string
		tokenPrice        types.TokenPrice
		expectedICQSubmit bool
	}{
		{
			name: "never updated token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "uatom",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   1,
				LastRequestTime: time.Time{}, // Zero time
			},
			expectedICQSubmit: true,
		},
		{
			name: "stale token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "uosmo",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   2,
				LastRequestTime: staleTime,
			},
			expectedICQSubmit: true,
		},
		{
			name: "fresh token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "ustrd",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   3,
				LastRequestTime: freshTime,
			},
			expectedICQSubmit: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset test state
			s.SetupTest()

			// Setup params
			s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

			// Reset mock IcqKeeper
			s.App.ICQOracleKeeper.IcqKeeper = mockICQKeeper
			submitICQCalled = false

			// Store token price
			s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tc.tokenPrice)

			// Set block time to now
			s.Ctx = s.Ctx.WithBlockTime(now)

			// Run BeginBlocker
			s.App.ICQOracleKeeper.BeginBlocker(s.Ctx)

			// Verify if ICQ was submitted as expected
			s.Require().Equal(tc.expectedICQSubmit, submitICQCalled,
				"ICQ submission status does not match expected for case: %s", tc.name)

			// If ICQ was expected to be submitted, verify the token price query in progress flag
			if tc.expectedICQSubmit {
				updatedPrice := s.MustGetTokenPrice(
					tc.tokenPrice.BaseDenom,
					tc.tokenPrice.QuoteDenom,
					tc.tokenPrice.OsmosisPoolId,
				)
				s.Require().True(updatedPrice.QueryInProgress,
					"query in progress should be true after BeginBlocker for case: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBeginBlockerICQErrors() {
	// Setup mock ICQ keeper that returns an error
	s.mockICQKeeper = MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			return fmt.Errorf("icq submit failed")
		},
	}
	s.App.ICQOracleKeeper.IcqKeeper = s.mockICQKeeper

	// Create token price that needs updating
	updateIntervalSec := uint64(60)
	tokenPrice := types.TokenPrice{
		BaseDenom:       "uatom",
		QuoteDenom:      "uusdc",
		OsmosisPoolId:   1,
		LastRequestTime: time.Time{}, // Zero time to trigger update
	}
	s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)

	// Run BeginBlocker - should log error but continue
	err := s.App.ICQOracleKeeper.RefreshTokenPrice(s.Ctx, tokenPrice, updateIntervalSec)
	s.Require().ErrorContains(err, "failed to submit Osmosis CL pool ICQ")

	// Verify token price query was not submitted
	updatedPrice := s.MustGetTokenPrice(
		tokenPrice.BaseDenom,
		tokenPrice.QuoteDenom,
		tokenPrice.OsmosisPoolId,
	)
	s.Require().False(updatedPrice.QueryInProgress,
		"query in progress should remain false when ICQ submission fails")
}

func (s *KeeperTestSuite) TestBeginBlockerMultipleTokens() {
	var submittedQueries int

	// Setup mock ICQ keeper to count submitted queries
	mockICQKeeper := MockICQKeeper{
		SubmitICQRequestFn: func(ctx sdk.Context, query icqtypes.Query, forceUnique bool) error {
			submittedQueries++
			return nil
		},
	}
	s.App.ICQOracleKeeper.IcqKeeper = mockICQKeeper

	// Set params
	params := types.Params{
		UpdateIntervalSec: 60,
	}
	s.App.ICQOracleKeeper.SetParams(s.Ctx, params)

	now := time.Now().UTC()
	staleTime := now.Add(-2 * time.Minute)

	// Create multiple token prices
	tokenPrices := []types.TokenPrice{
		{
			BaseDenom:       "uatom",
			QuoteDenom:      "uusdc",
			OsmosisPoolId:   1,
			LastRequestTime: staleTime,
			QueryInProgress: false,
		},
		{
			BaseDenom:       "uosmo",
			QuoteDenom:      "uusdc",
			OsmosisPoolId:   2,
			LastRequestTime: staleTime,
			QueryInProgress: false,
		},
		{
			BaseDenom:       "ustrd",
			QuoteDenom:      "uusdc",
			OsmosisPoolId:   3,
			LastRequestTime: s.Ctx.BlockTime(), // Should skip this one
			QueryInProgress: true,
		},
		{
			BaseDenom:       "ustrd",
			QuoteDenom:      "uusdc",
			OsmosisPoolId:   4,
			LastRequestTime: s.Ctx.BlockTime(), // Should skip this one
			QueryInProgress: false,
		},
	}

	// Store all token prices
	for _, tp := range tokenPrices {
		s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tp)
	}

	// Set block time
	s.Ctx = s.Ctx.WithBlockTime(now)

	// Run BeginBlocker
	s.App.ICQOracleKeeper.BeginBlocker(s.Ctx)

	// Verify number of submitted queries
	s.Require().Equal(2, submittedQueries,
		"expected 2 ICQ queries to be submitted (skipping the one in progress)")

	// Verify query in progress flags
	for _, tp := range tokenPrices[:2] {
		updatedPrice := s.MustGetTokenPrice(tp.BaseDenom, tp.QuoteDenom, tp.OsmosisPoolId)
		s.Require().True(updatedPrice.QueryInProgress,
			"query in progress should be set to true for tokens that are updating")
	}
	for _, tp := range tokenPrices[2:] {
		updatedPrice := s.MustGetTokenPrice(tp.BaseDenom, tp.QuoteDenom, tp.OsmosisPoolId)
		s.Require().Equal(tp.QueryInProgress, updatedPrice.QueryInProgress,
			"query in progress should not change for tokens that are not updating")
	}
}
