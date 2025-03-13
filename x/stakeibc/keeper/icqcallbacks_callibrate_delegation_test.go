package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	icqtypes "github.com/Stride-Labs/stride/v26/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestCalibrateDelegation_Success() {
	queriedValIndex := 1
	initialTotalDelegations := sdkmath.NewInt(1_000_000)

	baseHostZone := types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: initialTotalDelegations,
		Validators: []*types.Validator{
			{Address: "valoper1"}, // not queried
			{Address: ValAddress}, // queried validator - will get overridden in each test case
		},
	}

	testCases := []struct {
		name                  string
		currentDelegation     sdkmath.Int
		sharesInQueryResponse sdkmath.LegacyDec
		sharesToTokensRate    sdkmath.LegacyDec
		expectedEndDelegation sdkmath.Int
	}{
		{
			// Current delegation: 10,000 tokens
			// Query response:     13,334 shares * 0.75 sharesToTokens = 10,000 tokens (+0)
			name:                  "delegation change of 0",
			currentDelegation:     sdkmath.NewInt(10_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("13334"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(10_000),
		},
		{
			// Current delegation: 10,000 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-2,500)
			name:                  "negative delegation change",
			currentDelegation:     sdkmath.NewInt(10_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 12,500 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+2,500)
			name:                  "positive delegation change",
			currentDelegation:     sdkmath.NewInt(12_500),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 12,500 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-5,000)
			name:                  "negative delegation change at threshold boundary",
			currentDelegation:     sdkmath.NewInt(12_500),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 10,000 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,000)
			name:                  "positive delegation change at threshold boundary",
			currentDelegation:     sdkmath.NewInt(10_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 12,501 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-5,001)
			name:                  "negative delegation change exceeds threshold",
			currentDelegation:     sdkmath.NewInt(12_501),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(12_501), // no change
		},
		{
			// Current delegation: 9,999 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,001)
			name:                  "positive delegation change exceeds threshold",
			currentDelegation:     sdkmath.NewInt(9_999),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(9_999), // no change
		},
	}

	for _, tc := range testCases {
		// Define a host zone with the current parameters
		hostZone := baseHostZone
		hostZone.Validators[queriedValIndex] = &types.Validator{
			Address:            ValAddress,
			Delegation:         tc.currentDelegation,
			SharesToTokensRate: tc.sharesToTokensRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

		// Mock out the query response and confirm the callback succeede
		query := icqtypes.Query{ChainId: HostChainId}
		queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, tc.sharesInQueryResponse)

		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
		s.Require().NoError(err, "%s - no error expected during delegation callback", tc.name)

		// Fetch the updated host zone and validator
		updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
		s.Require().True(found, "%s - host zone should have been found", tc.name)
		updatedValidator := updatedHostZone.Validators[queriedValIndex]

		// Confirm the delegation changes match expectations
		expectedDelegationChange := tc.expectedEndDelegation.Sub(tc.currentDelegation)
		expectedTotalDelegation := initialTotalDelegations.Add(expectedDelegationChange)
		s.Require().Equal(tc.expectedEndDelegation.Int64(), updatedValidator.Delegation.Int64(),
			"%s - validator delegation", tc.name)
		s.Require().Equal(expectedTotalDelegation.Int64(), updatedHostZone.TotalDelegations.Int64(),
			"%s - host zone total delegation", tc.name)
	}
}

func (s *KeeperTestSuite) TestCalibrateDelegation_Failure() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:    HostChainId,
		Validators: []*types.Validator{{Address: ValAddress}},
	})
	validQuery := icqtypes.Query{ChainId: HostChainId}
	validQueryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, sdkmath.LegacyNewDec(1000))

	// Atempt the callback with a missing host zone - it should fail
	invalidQuery := validQuery
	invalidQuery.ChainId = ""
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, validQueryResponse, invalidQuery)
	s.Require().ErrorContains(err, "host zone not found")

	// Attempt the callback with an invalid query response - it should fail
	invalidQueryResponse := []byte{1, 2, 3}
	err = keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, invalidQueryResponse, validQuery)
	s.Require().ErrorContains(err, "unable to unmarshal delegator shares query response")

	// Attempt the callback with a non-existent validator address - it should fail
	invalidQueryResponse = s.CreateDelegatorSharesQueryResponse("non-existent validator", sdkmath.LegacyNewDec(1000))
	err = keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, invalidQueryResponse, validQuery)
	s.Require().ErrorContains(err, "validator not found")
}
