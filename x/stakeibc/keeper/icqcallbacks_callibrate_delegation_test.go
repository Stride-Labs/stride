package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
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
			name:                  "large negative delegation change",
			currentDelegation:     sdkmath.NewInt(12_500),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 10,000 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,000)
			name:                  "large positive delegation change",
			currentDelegation:     sdkmath.NewInt(10_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 12,501 tokens
			// Query response:     10,000 shares * 0.75 sharesToTokens = 7,500 tokens (-5,001)
			// There is no longer a cap on the change (wind-down spec §5), so this applies too
			name:                  "negative delegation change above the former cap",
			currentDelegation:     sdkmath.NewInt(12_501),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("10000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(7_500),
		},
		{
			// Current delegation: 9,999 tokens
			// Query response:     20,000 shares * 0.75 sharesToTokens = 15,000 tokens (+5,001)
			name:                  "positive delegation change above the former cap",
			currentDelegation:     sdkmath.NewInt(9_999),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("20000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.75"),
			expectedEndDelegation: sdkmath.NewInt(15_000),
		},
		{
			// Current delegation: 1,000,000,000 tokens (the whole zone)
			// Query response:     1,000,000,000 shares * 0.5 sharesToTokens = 500,000,000 tokens (-500,000,000)
			name:                  "delegation change of half the zone",
			currentDelegation:     sdkmath.NewInt(1_000_000_000),
			sharesInQueryResponse: sdkmath.LegacyMustNewDecFromStr("1000000000"),
			sharesToTokensRate:    sdkmath.LegacyMustNewDecFromStr("0.5"),
			expectedEndDelegation: sdkmath.NewInt(500_000_000),
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

// The callback must leave state untouched when the queried shares can't be trusted
func (s *KeeperTestSuite) TestCalibrateDelegation_NoOp() {
	initialDelegation := sdkmath.NewInt(10_000)
	initialTotalDelegations := sdkmath.NewInt(1_000_000)

	testCases := []struct {
		name                        string
		delegationChangesInProgress int64
		sharesToTokensRate          sdkmath.LegacyDec
	}{
		{
			// A delegation ICA is in flight, so the queried shares race the recorded delegation
			name:                        "delegation change in progress",
			delegationChangesInProgress: 1,
			sharesToTokensRate:          sdkmath.LegacyMustNewDecFromStr("0.75"),
		},
		{
			// A zero rate would compute zero tokens and wipe the recorded delegation
			name:               "zero shares to tokens rate",
			sharesToTokensRate: sdkmath.LegacyZeroDec(),
		},
		{
			// A nil rate is unset and would compute a meaningless token amount
			name: "nil shares to tokens rate",
		},
	}

	for _, tc := range testCases {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
			ChainId:          HostChainId,
			TotalDelegations: initialTotalDelegations,
			Validators: []*types.Validator{{
				Address:                     ValAddress,
				Delegation:                  initialDelegation,
				SharesToTokensRate:          tc.sharesToTokensRate,
				DelegationChangesInProgress: tc.delegationChangesInProgress,
			}},
		})

		query := icqtypes.Query{ChainId: HostChainId}
		queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, sdkmath.LegacyMustNewDecFromStr("10000"))

		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
		s.Require().NoError(err, "%s - no error expected", tc.name)

		updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
		s.Require().True(found, "%s - host zone should have been found", tc.name)
		s.Require().Equal(initialDelegation.Int64(), updatedHostZone.Validators[0].Delegation.Int64(),
			"%s - validator delegation should be unchanged", tc.name)
		s.Require().Equal(initialTotalDelegations.Int64(), updatedHostZone.TotalDelegations.Int64(),
			"%s - host zone total delegation should be unchanged", tc.name)
	}
}

// An empty response means the delegation ICA has no delegation to the validator on the host,
// so the recorded delegation is corrected to zero (whatever the stored rate is)
func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponse() {
	initialTotalDelegations := sdkmath.NewInt(1_000_000)

	for _, rate := range []sdkmath.LegacyDec{sdkmath.LegacyMustNewDecFromStr("0.75"), sdkmath.LegacyZeroDec(), {}} {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
			ChainId:          HostChainId,
			TotalDelegations: initialTotalDelegations,
			Validators: []*types.Validator{
				{Address: "valoper1", Delegation: sdkmath.NewInt(500)},
				{Address: ValAddress, Delegation: sdkmath.NewInt(10_000), SharesToTokensRate: rate},
			},
		})

		query := icqtypes.Query{ChainId: HostChainId, CallbackData: []byte(ValAddress)}
		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, query)
		s.Require().NoError(err, "rate %v", rate)

		hostZone := s.MustGetHostZone(HostChainId)
		s.Require().Equal(int64(500), hostZone.Validators[0].Delegation.Int64(), "other validator untouched")
		s.Require().Equal(int64(0), hostZone.Validators[1].Delegation.Int64(), "validator delegation, rate %v", rate)
		s.Require().Equal(int64(990_000), hostZone.TotalDelegations.Int64(), "total delegations, rate %v", rate)
	}
}

func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponse_DelegationChangeInProgress() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: sdkmath.NewInt(1_000_000),
		Validators: []*types.Validator{{
			Address:                     ValAddress,
			Delegation:                  sdkmath.NewInt(10_000),
			SharesToTokensRate:          sdkmath.LegacyMustNewDecFromStr("0.75"),
			DelegationChangesInProgress: 1,
		}},
	})

	query := icqtypes.Query{ChainId: HostChainId, CallbackData: []byte(ValAddress)}
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, query)
	s.Require().NoError(err)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(10_000), hostZone.Validators[0].Delegation.Int64())
	s.Require().Equal(int64(1_000_000), hostZone.TotalDelegations.Int64())
}

func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponse_Failure() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:    HostChainId,
		Validators: []*types.Validator{{Address: ValAddress}},
	})

	// No validator address in the callback data
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, icqtypes.Query{ChainId: HostChainId})
	s.Require().ErrorContains(err, "callback data")

	// Unknown validator address
	query := icqtypes.Query{ChainId: HostChainId, CallbackData: []byte("non-existent validator")}
	err = keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, query)
	s.Require().ErrorContains(err, "validator not found")
}
