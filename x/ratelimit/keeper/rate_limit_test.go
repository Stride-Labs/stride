package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	minttypes "github.com/Stride-Labs/stride/v5/x/mint/types"
	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

const (
	denom     = "denom"
	channelId = "channel-0"
)

type action struct {
	direction           types.PacketDirection
	amount              int64
	addToBlacklist      bool
	removeFromBlacklist bool
}

type checkRateLimitTestCase struct {
	name          string
	actions       []action
	expectedError string
}

// Helper function to check if an element is in an array
func isInArray(element string, arr []string) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}

func (s *KeeperTestSuite) TestGetChannelValue() {
	supply := sdkmath.NewInt(100)

	// Mint coins to increase the supply, which will increase the channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, supply)))
	s.Require().NoError(err)

	expected := supply
	actual := s.App.RatelimitKeeper.GetChannelValue(s.Ctx, denom)
	s.Require().Equal(expected, actual)
}

// Helper function to create 5 rate limit objects with various attributes
func (s *KeeperTestSuite) createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		rateLimit := types.RateLimit{
			Path: &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
			Flow: &types.Flow{Inflow: sdkmath.NewInt(10), Outflow: sdkmath.NewInt(10)},
		}

		rateLimits = append(rateLimits, rateLimit)
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestGetRateLimit() {
	rateLimits := s.createRateLimits()

	expectedRateLimit := rateLimits[0]
	denom := expectedRateLimit.Path.Denom
	channelId := expectedRateLimit.Path.ChannelId

	actualRateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
	s.Require().True(found, "element should have been found, but was not")
	s.Require().Equal(expectedRateLimit, actualRateLimit)
}

func (s *KeeperTestSuite) TestRemoveRateLimit() {
	rateLimits := s.createRateLimits()

	rateLimitToRemove := rateLimits[0]
	denomToRemove := rateLimitToRemove.Path.Denom
	channelIdToRemove := rateLimitToRemove.Path.ChannelId

	s.App.RatelimitKeeper.RemoveRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	_, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	s.Require().False(found, "the removed element should not have been found, but it was")
}

func (s *KeeperTestSuite) TestResetRateLimit() {
	rateLimits := s.createRateLimits()

	rateLimitToReset := rateLimits[0]
	denomToRemove := rateLimitToReset.Path.Denom
	channelIdToRemove := rateLimitToReset.Path.ChannelId

	err := s.App.RatelimitKeeper.ResetRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	s.Require().NoError(err)

	rateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denomToRemove, channelIdToRemove)
	s.Require().True(found, "element should have been found, but was not")
	s.Require().Zero(rateLimit.Flow.Inflow.Int64(), "Inflow should have been reset to 0")
	s.Require().Zero(rateLimit.Flow.Outflow.Int64(), "Outflow should have been reset to 0")
}

func (s *KeeperTestSuite) TestGetAllRateLimits() {
	expectedRateLimits := s.createRateLimits()
	actualRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Len(actualRateLimits, len(expectedRateLimits))
	s.Require().ElementsMatch(expectedRateLimits, actualRateLimits, "all rate limits")
}

func (s *KeeperTestSuite) TestDenomBlacklist() {
	allDenoms := []string{"denom1", "denom2", "denom3", "denom4"}
	denomsToBlacklist := []string{"denom1", "denom3"}

	// No denoms are currently blacklisted
	for _, denom := range allDenoms {
		isBlacklisted := s.App.RatelimitKeeper.IsDenomBlacklisted(s.Ctx, denom)
		s.Require().False(isBlacklisted, "%s should not be blacklisted yet", denom)
	}

	// Blacklist two denoms
	for _, denom := range denomsToBlacklist {
		s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, denom)
	}

	// Confirm half the list was blacklisted and the others were not
	for _, denom := range allDenoms {
		isBlacklisted := s.App.RatelimitKeeper.IsDenomBlacklisted(s.Ctx, denom)

		if isInArray(denom, denomsToBlacklist) {
			s.Require().True(isBlacklisted, "%s should have been blacklisted", denom)
		} else {
			s.Require().False(isBlacklisted, "%s should not have been blacklisted", denom)
		}
	}
	actualBlacklistedDenoms := s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx)
	s.Require().Len(actualBlacklistedDenoms, len(denomsToBlacklist), "number of blacklisted denoms")
	s.Require().ElementsMatch(denomsToBlacklist, actualBlacklistedDenoms, "list of blacklisted denoms")

	// Finally, remove denoms from blacklist and confirm they were removed
	for _, denom := range denomsToBlacklist {
		s.App.RatelimitKeeper.RemoveDenomFromBlacklist(s.Ctx, denom)
	}
	for _, denom := range allDenoms {
		isBlacklisted := s.App.RatelimitKeeper.IsDenomBlacklisted(s.Ctx, denom)

		if isInArray(denom, denomsToBlacklist) {
			s.Require().False(isBlacklisted, "%s should have been removed from the blacklist", denom)
		} else {
			s.Require().False(isBlacklisted, "%s should never have been blacklisted", denom)
		}
	}
}

// Adds a rate limit object to the store in preparation for the check rate limit tests
func (s *KeeperTestSuite) SetupCheckRateLimitAndUpdateFlowTest() {
	channelValue := sdkmath.NewInt(100)
	maxPercentSend := sdkmath.NewInt(10)
	maxPercentRecv := sdkmath.NewInt(10)

	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
		Path: &types.Path{
			Denom:     denom,
			ChannelId: channelId,
		},
		Quota: &types.Quota{
			MaxPercentSend: maxPercentSend,
			MaxPercentRecv: maxPercentRecv,
			DurationHours:  1,
		},
		Flow: &types.Flow{
			Inflow:       sdkmath.ZeroInt(),
			Outflow:      sdkmath.ZeroInt(),
			ChannelValue: channelValue,
		},
	})

	s.App.RatelimitKeeper.RemoveDenomFromBlacklist(s.Ctx, denom)
}

// Helper function to check the rate limit across a series of transfers
func (s *KeeperTestSuite) processCheckRateLimitAndUpdateFlowTestCase(tc checkRateLimitTestCase) {
	s.SetupCheckRateLimitAndUpdateFlowTest()

	expectedInflow := sdkmath.NewInt(0)
	expectedOutflow := sdkmath.NewInt(0)
	for i, action := range tc.actions {
		if action.addToBlacklist {
			s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, denom)
			continue
		} else if action.removeFromBlacklist {
			s.App.RatelimitKeeper.RemoveDenomFromBlacklist(s.Ctx, denom)
			continue
		}

		amount := sdkmath.NewInt(action.amount)
		err := s.App.RatelimitKeeper.CheckRateLimitAndUpdateFlow(s.Ctx, action.direction, denom, channelId, amount)

		// Only check the error on the last action
		if i == len(tc.actions)-1 && tc.expectedError != "" {
			s.Require().ErrorContains(err, tc.expectedError, tc.name+"- action: #%d - error", i)
		} else {
			// All but the last action should succeed
			s.Require().NoError(err, tc.name+"- action: #%d - no error", i)

			// Update expected flow
			if action.direction == types.PACKET_RECV {
				expectedInflow = expectedInflow.Add(amount)
			} else {
				expectedOutflow = expectedOutflow.Add(amount)
			}
		}

		// Confirm flow is updated properly (or left as is if the theshold was exceeded)
		rateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found)
		s.Require().Equal(expectedInflow, rateLimit.Flow.Inflow, tc.name+"- action: #%d - inflow", i)
		s.Require().Equal(expectedOutflow, rateLimit.Flow.Outflow, tc.name+"- action: #%d - outflow", i)
	}
}

func (s *KeeperTestSuite) TestCheckRateLimitAndUpdateFlow_UnidirectionalFlow() {
	testCases := []checkRateLimitTestCase{
		{
			name: "send_under_threshold",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 5},
				{direction: types.PACKET_SEND, amount: 5},
			},
		},
		{
			name: "send_over_threshold",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 5},
				{direction: types.PACKET_SEND, amount: 6},
			},
			expectedError: "Outflow exceeds quota",
		},
		{
			name: "recv_under_threshold",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 5},
				{direction: types.PACKET_RECV, amount: 5},
			},
		},
		{
			name: "recv_over_threshold",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 5},
				{direction: types.PACKET_RECV, amount: 6},
			},
			expectedError: "Inflow exceeds quota",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestCheckRateLimitAndUpdatedFlow_BidirectionalFlow() {
	testCases := []checkRateLimitTestCase{
		{
			name: "send_then_recv_under_threshold",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
			},
		},
		{
			name: "recv_then_send_under_threshold",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
			},
		},
		{
			name: "send_then_recv_over_inflow",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: -2
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +4
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +2
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +8
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +6
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +12 (exceeds threshold)
			},
			expectedError: "Inflow exceeds quota",
		},
		{
			name: "send_then_recv_over_outflow",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6}, //   -6, Net: -6
				{direction: types.PACKET_RECV, amount: 2}, //   +2, Net: -4
				{direction: types.PACKET_SEND, amount: 6}, //   -6, Net: -10
				{direction: types.PACKET_RECV, amount: 2}, //   +2, Net: -8
				{direction: types.PACKET_SEND, amount: 6}, //   -6, Net: -14 (exceeds threshold)
			},
			expectedError: "Outflow exceeds quota",
		},
		{
			name: "recv_then_send_over_inflow",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +6
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +4
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +10
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +8
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +14 (exceeds threshold)
			},
			expectedError: "Inflow exceeds quota",
		},
		{
			name: "recv_then_send_over_outflow",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 2},  //  +2, Net: +2
				{direction: types.PACKET_SEND, amount: 6},  //  -6, Net: -4
				{direction: types.PACKET_RECV, amount: 2},  //  +2, Net: -2
				{direction: types.PACKET_SEND, amount: 6},  //  -6, Net: -8
				{direction: types.PACKET_RECV, amount: 2},  //  +2, Net: -6
				{direction: types.PACKET_SEND, amount: 10}, //  +6, Net: -12 (exceeds threshold)
			},
			expectedError: "Outflow exceeds quota",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestCheckRateLimitAndUpdatedFlow_Blacklist() {
	testCases := []checkRateLimitTestCase{
		{
			name: "add_then_remove_from_blacklist", // should succeed
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToBlacklist: true},
				{removeFromBlacklist: true},
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
			},
		},
		{
			name: "send_recv_blacklist_send",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_SEND, amount: 6},
			},
			expectedError: types.ErrDenomIsBlacklisted.Error(),
		},
		{
			name: "send_recv_blacklist_recv",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_RECV, amount: 6},
			},
			expectedError: types.ErrDenomIsBlacklisted.Error(),
		},
		{
			name: "recv_send_blacklist_send",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_SEND, amount: 6},
			},
			expectedError: types.ErrDenomIsBlacklisted.Error(),
		},
		{
			name: "recv_send_blacklist_recv",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_RECV, amount: 6},
			},
			expectedError: types.ErrDenomIsBlacklisted.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}
