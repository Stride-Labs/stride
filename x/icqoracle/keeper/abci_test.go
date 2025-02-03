package keeper_test

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

func (s *KeeperTestSuite) TestBeginBlockerParams() {
	// Delete params from store
	s.DeleteParams()

	// Run BeginBlocker with missing params
	s.App.ICQOracleKeeper.BeginBlocker(s.Ctx)

	// Get the logged output
	logOutput := s.logBuffer.String()

	// Verify the error was logged
	s.Require().True(
		strings.Contains(logOutput, "failed to get icqoracle params"),
		"expected error log message about missing params, got: %s", logOutput,
	)
}

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
				BaseDenom:     "uatom",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "1",
				UpdatedAt:     time.Time{}, // Zero time
			},
			expectedICQSubmit: true,
		},
		{
			name: "stale token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:     "uosmo",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "2",
				UpdatedAt:     staleTime,
			},
			expectedICQSubmit: true,
		},
		{
			name: "fresh token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:     "ustrd",
				QuoteDenom:    "uusdc",
				OsmosisPoolId: "3",
				UpdatedAt:     freshTime,
			},
			expectedICQSubmit: false,
		},
		{
			name: "query in progress stale token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "ujuno",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   "4",
				UpdatedAt:       staleTime,
				QueryInProgress: true,
			},
			expectedICQSubmit: false,
		},
		{
			name: "query in progress fresh token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "udydx",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   "5",
				UpdatedAt:       freshTime,
				QueryInProgress: true,
			},
			expectedICQSubmit: false,
		},
		{
			name: "query in progress never updated token price",
			tokenPrice: types.TokenPrice{
				BaseDenom:       "utia",
				QuoteDenom:      "uusdc",
				OsmosisPoolId:   "6",
				UpdatedAt:       time.Time{}, // Zero time
				QueryInProgress: true,
			},
			expectedICQSubmit: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset test state
			s.SetupTest()

			// Setup params
			err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
			s.Require().NoError(err)

			// Reset mock IcqKeeper
			s.App.ICQOracleKeeper.IcqKeeper = mockICQKeeper
			submitICQCalled = false

			// Store token price
			err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tc.tokenPrice)
			s.Require().NoError(err)

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

	// Set params
	params := types.Params{
		UpdateIntervalSec: 60,
	}
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	// Create token price that needs updating
	tokenPrice := types.TokenPrice{
		BaseDenom:     "uatom",
		QuoteDenom:    "uusdc",
		OsmosisPoolId: "1",
		UpdatedAt:     time.Time{}, // Zero time to trigger update
	}
	err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tokenPrice)
	s.Require().NoError(err)

	// Run BeginBlocker - should log error but continue
	s.App.ICQOracleKeeper.BeginBlocker(s.Ctx)

	// Get the logged output and verify error was logged
	logOutput := s.logBuffer.String()
	s.Require().True(
		strings.Contains(logOutput, "icq submit failed"),
		"expected error log message about ICQ submission failure, got: %s", logOutput,
	)

	// Verify token price was not modified
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
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err)

	now := time.Now().UTC()
	staleTime := now.Add(-2 * time.Minute)

	// Create multiple token prices
	tokenPrices := []types.TokenPrice{
		{
			BaseDenom:     "uatom",
			QuoteDenom:    "uusdc",
			OsmosisPoolId: "1",
			UpdatedAt:     staleTime,
		},
		{
			BaseDenom:     "uosmo",
			QuoteDenom:    "uusdc",
			OsmosisPoolId: "2",
			UpdatedAt:     staleTime,
		},
		{
			BaseDenom:       "ustrd",
			QuoteDenom:      "uusdc",
			OsmosisPoolId:   "3",
			UpdatedAt:       staleTime,
			QueryInProgress: true, // Should skip this one
		},
	}

	// Store all token prices
	for _, tp := range tokenPrices {
		err = s.App.ICQOracleKeeper.SetTokenPrice(s.Ctx, tp)
		s.Require().NoError(err)
	}

	// Set block time
	s.Ctx = s.Ctx.WithBlockTime(now)

	// Run BeginBlocker
	s.App.ICQOracleKeeper.BeginBlocker(s.Ctx)

	// Verify number of submitted queries
	s.Require().Equal(2, submittedQueries,
		"expected 2 ICQ queries to be submitted (skipping the one in progress)")

	// Verify query in progress flags
	for _, tp := range tokenPrices {
		updatedPrice := s.MustGetTokenPrice(tp.BaseDenom, tp.QuoteDenom, tp.OsmosisPoolId)
		if tp.QueryInProgress {
			s.Require().True(updatedPrice.QueryInProgress,
				"query in progress should remain true for token that was already updating")
		} else {
			s.Require().True(updatedPrice.QueryInProgress,
				"query in progress should be true for tokens that needed updates")
		}
	}
}
