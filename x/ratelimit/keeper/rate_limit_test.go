package keeper_test

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	minttypes "github.com/Stride-Labs/stride/v4/x/mint/types"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

const (
	denom     = "denom"
	channelId = "channel-0"
)

type action struct {
	direction types.PacketDirection
	amount    int64
}

type checkRateLimitTestCase struct {
	name          string
	actions       []action
	expectedError string
}

func (s *KeeperTestSuite) TestGetChannelValue() {
	supply := sdk.NewInt(100)

	// Mint coins to increase the supply, which will increase the channel value
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, supply)))
	s.Require().NoError(err)

	expected := supply
	actual := s.App.RatelimitKeeper.GetChannelValue(s.Ctx, denom)
	s.Require().Equal(expected, actual)
}

func (s *KeeperTestSuite) createRateLimits() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := 1; i <= 5; i++ {
		suffix := strconv.Itoa(i)
		rateLimit := types.RateLimit{
			Path: &types.Path{Denom: "denom-" + suffix, ChannelId: "channel-" + suffix},
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

func (s *KeeperTestSuite) TestGetAllRateLimits() {
	expectedRateLimits := s.createRateLimits()
	actualRateLimits := s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx)
	s.Require().Len(actualRateLimits, len(expectedRateLimits))
	s.Require().ElementsMatch(expectedRateLimits, actualRateLimits, "all rate limits")
}

func (s *KeeperTestSuite) SetupCheckRateLimitTest() {
	// Add rate limit to store
	channelValue := sdk.NewInt(100)
	maxPercentSend := sdk.NewInt(10)
	maxPercentRecv := sdk.NewInt(10)

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
			Inflow:       sdk.ZeroInt(),
			Outflow:      sdk.ZeroInt(),
			ChannelValue: channelValue,
		},
	})
}

func (s *KeeperTestSuite) ProcessCheckRateLimitTestCase(tc checkRateLimitTestCase) {
	s.SetupCheckRateLimitTest()

	expectedInflow := sdk.NewInt(0)
	expectedOutflow := sdk.NewInt(0)
	for i, action := range tc.actions {
		amount := sdk.NewInt(action.amount)
		err := s.App.RatelimitKeeper.CheckRateLimitAndUpdateFlow(s.Ctx, action.direction, denom, channelId, amount)

		if i == len(tc.actions)-1 && tc.expectedError != "" {
			s.Require().ErrorIs(err, types.ErrQuotaExceeded, tc.name+" - action: #%d - error type", i)
			s.Require().ErrorContains(err, tc.expectedError, tc.name+"- action: #%d - error string", i)
		} else {
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

func (s *KeeperTestSuite) TestCheckRateLimit_UnilateralFlow() {
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
			s.ProcessCheckRateLimitTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestCheckRateLimit_BidirectionalFlow() {
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
			s.ProcessCheckRateLimitTestCase(tc)
		})
	}
}
