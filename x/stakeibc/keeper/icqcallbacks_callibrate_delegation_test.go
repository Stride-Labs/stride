package keeper_test

import (
	"time"

	"github.com/cosmos/gogoproto/proto"

	sdkmath "cosmossdk.io/math"

	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
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
			name:                  "zero recorded delegation snapshot remains usable",
			currentDelegation:     sdkmath.ZeroInt(),
			sharesInQueryResponse: sdkmath.LegacyNewDec(20_000),
			sharesToTokensRate:    sdkmath.LegacyOneDec(),
			expectedEndDelegation: sdkmath.NewInt(20_000),
		},
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

		// Snapshot the recorded delegation when the query is submitted.
		query := s.calibrationQuery(tc.currentDelegation)
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
	validQuery := s.calibrationQuery(sdkmath.ZeroInt())
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
		{
			name:               "negative shares to tokens rate",
			sharesToTokensRate: sdkmath.LegacyNewDec(-1),
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

		query := s.calibrationQuery(initialDelegation)
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

// The calibration query as SubmitCalibrationICQ stores it, for ValAddress
func (s *KeeperTestSuite) calibrationQuery(initialDelegation sdkmath.Int) icqtypes.Query {
	callbackData, err := proto.Marshal(&types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: initialDelegation,
		ValidatorAddress:           ValAddress,
	})
	s.Require().NoError(err)
	return icqtypes.Query{ChainId: HostChainId, CallbackData: callbackData, InvokeCallbackOnEmptyResponse: true}
}

// An empty response proves the delegation ICA has no delegation to the validator on the host:
// the recorded delegation is corrected to zero, with or without a usable stored rate
func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponseCorrectsToZero() {
	recordedDelegation := sdkmath.NewInt(7_500)
	otherDelegation := sdkmath.NewInt(2_000)
	initialTotalDelegations := sdkmath.NewInt(1_000_000)

	testCases := []struct {
		name               string
		sharesToTokensRate sdkmath.LegacyDec
	}{
		{name: "usable rate", sharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.9")},
		{name: "zero rate", sharesToTokensRate: sdkmath.LegacyZeroDec()},
		{name: "nil rate"},
	}

	for _, tc := range testCases {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
			ChainId:          HostChainId,
			TotalDelegations: initialTotalDelegations,
			Validators: []*types.Validator{
				{Address: "other_validator", Delegation: otherDelegation, SharesToTokensRate: sdkmath.LegacyOneDec()},
				{Address: ValAddress, Delegation: recordedDelegation, SharesToTokensRate: tc.sharesToTokensRate},
			},
		})

		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, s.calibrationQuery(recordedDelegation))
		s.Require().NoError(err, "%s - no error expected", tc.name)

		hostZone := s.MustGetHostZone(HostChainId)
		s.Require().Equal(otherDelegation, hostZone.Validators[0].Delegation, "%s - other validator untouched", tc.name)
		s.Require().Equal(sdkmath.ZeroInt(), hostZone.Validators[1].Delegation, "%s - delegation corrected to zero", tc.name)
		s.Require().Equal(initialTotalDelegations.Sub(recordedDelegation), hostZone.TotalDelegations,
			"%s - total delegations lowered by the phantom amount", tc.name)
	}
}

// Through the interchain-query msg server: an empty response to a stored calibration query
// reaches the callback, which zeroes the recorded delegation
func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyQueryResponseReachesCallback() {
	recordedDelegation := sdkmath.NewInt(7_500)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: sdkmath.NewInt(10_000),
		Validators: []*types.Validator{{
			Address: ValAddress, Delegation: recordedDelegation, SharesToTokensRate: sdkmath.LegacyOneDec(),
		}},
	})

	query := s.calibrationQuery(recordedDelegation)
	query.Id = "calibration-query"
	query.CallbackModule = types.ModuleName
	query.CallbackId = keeper.ICQCallbackID_Calibrate
	// No "key" suffix, so no proof is verified. SubmitICQRequest rejects that combination with
	// the opt-in; the query is stored directly here because the test has no host proof to offer
	query.QueryType = "store/staking"
	query.TimeoutTimestamp = uint64(s.Ctx.BlockTime().Add(time.Hour).UnixNano())
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)

	msgServer := icqkeeper.NewMsgServerImpl(s.App.InterchainqueryKeeper)
	_, err := msgServer.SubmitQueryResponse(s.Ctx, &icqtypes.MsgSubmitQueryResponse{
		ChainId:     HostChainId,
		QueryId:     query.Id,
		Result:      []byte{},
		FromAddress: s.TestAccs[0].String(),
	})
	s.Require().NoError(err)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(sdkmath.ZeroInt(), hostZone.Validators[0].Delegation, "delegation corrected to zero")
	s.Require().Equal(sdkmath.NewInt(2_500), hostZone.TotalDelegations, "total delegations lowered by the phantom amount")
	_, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, query.Id)
	s.Require().False(found, "query deleted")
}

// An empty response is still subject to the checks that protect a non-empty one
func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponseNoOp() {
	recordedDelegation := sdkmath.NewInt(7_500)
	initialTotalDelegations := sdkmath.NewInt(1_000_000)

	testCases := []struct {
		name                        string
		snapshotDelegation          sdkmath.Int
		delegationChangesInProgress int64
	}{
		{
			// An undelegate ICA is in flight, so the proven absence may predate it
			name:                        "delegation change in progress",
			snapshotDelegation:          recordedDelegation,
			delegationChangesInProgress: 1,
		},
		{
			// The recorded delegation moved after the query was submitted
			name:               "delegation changed since the snapshot",
			snapshotDelegation: sdkmath.NewInt(9_000),
		},
	}

	for _, tc := range testCases {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
			ChainId:          HostChainId,
			TotalDelegations: initialTotalDelegations,
			Validators: []*types.Validator{{
				Address:                     ValAddress,
				Delegation:                  recordedDelegation,
				SharesToTokensRate:          sdkmath.LegacyOneDec(),
				DelegationChangesInProgress: tc.delegationChangesInProgress,
			}},
		})
		before := s.MustGetHostZone(HostChainId)

		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, s.calibrationQuery(tc.snapshotDelegation))
		s.Require().NoError(err, "%s - no error expected", tc.name)
		s.Require().Equal(before, s.MustGetHostZone(HostChainId), "%s - accounting unchanged", tc.name)
	}
}

// The validator of an empty response comes from the callback data; one the zone does not
// track is an error, as it is for a non-empty response
func (s *KeeperTestSuite) TestCalibrateDelegation_EmptyResponseUnknownValidator() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: sdkmath.NewInt(5_000),
		Validators:       []*types.Validator{{Address: "other_validator", Delegation: sdkmath.NewInt(5_000)}},
	})
	before := s.MustGetHostZone(HostChainId)

	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, []byte{}, s.calibrationQuery(sdkmath.NewInt(5_000)))
	s.Require().ErrorIs(err, types.ErrValidatorNotFound)
	s.Require().Equal(before, s.MustGetHostZone(HostChainId), "accounting unchanged")
}

func (s *KeeperTestSuite) TestCalibrateDelegation_CompletedUndelegation() {
	tc := s.SetupUndelegateCallbackNoRecords()
	hostZone := s.MustGetHostZone(HostChainId)
	hostZone.Validators[0].SharesToTokensRate = sdkmath.LegacyOneDec()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	query := s.calibrationQuery(hostZone.Validators[0].Delegation)
	queryResponse := s.CreateDelegatorSharesQueryResponse(hostZone.Validators[0].Address,
		sdkmath.LegacyNewDecFromInt(hostZone.Validators[0].Delegation))

	// Complete the real ICA callback before the older calibration answer arrives.
	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().NoError(err)
	afterUndelegation := s.MustGetHostZone(HostChainId)
	s.Require().Zero(afterUndelegation.Validators[0].DelegationChangesInProgress)
	s.Require().Equal(tc.initialState.val1Bal.Sub(tc.val1UndelegationAmount), afterUndelegation.Validators[0].Delegation)

	err = keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
	s.Require().NoError(err)
	s.Require().Equal(afterUndelegation, s.MustGetHostZone(HostChainId), "stale shares must not resurrect undelegated tokens")
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "stale calibration does not start a retry loop")
}

func (s *KeeperTestSuite) TestCalibrateDelegation_InvalidCallbackData() {
	for _, callbackData := range [][]byte{nil, {}, {1, 2, 3}, {0x10, 0x01}} {
		hostZone := types.HostZone{
			ChainId: HostChainId, TotalDelegations: sdkmath.NewInt(10_000),
			Validators: []*types.Validator{{
				Address: ValAddress, Delegation: sdkmath.NewInt(10_000),
				SharesToTokensRate: sdkmath.LegacyOneDec(),
			}},
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
		before := s.MustGetHostZone(HostChainId)
		queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, sdkmath.LegacyNewDec(20_000))
		err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse,
			icqtypes.Query{ChainId: HostChainId, CallbackData: callbackData})
		s.Require().NoError(err)
		s.Require().Equal(before, s.MustGetHostZone(HostChainId), "missing or malformed snapshot must not change accounting")
		s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx))
	}
}
