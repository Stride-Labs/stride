package keeper_test

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	minttypes "github.com/Stride-Labs/stride/v14/x/mint/types"
	"github.com/Stride-Labs/stride/v14/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v14/x/ratelimit/types"
)

const (
	denom     = "denom"
	channelId = "channel-0"
	sender    = "sender"
	receiver  = "receiver"
)

type action struct {
	direction           types.PacketDirection
	amount              int64
	addToBlacklist      bool
	removeFromBlacklist bool
	addToWhitelist      bool
	removeFromWhitelist bool
	skipFlowUpdate      bool
	expectedError       string
}

type checkRateLimitTestCase struct {
	name    string
	actions []action
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

func (s *KeeperTestSuite) TestPendingSendPacketPrefix() {
	// Store 5 packets across two channels
	sendPackets := []string{}
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, sequence)
			sendPackets = append(sendPackets, fmt.Sprintf("%s/%d", channelId, sequence))
		}
	}

	// Check that they each sequence number is found
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			found := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().True(found, "send packet should have been found - channel %s, sequence: %d", channelId, sequence)
		}
	}

	// Check lookup of all sequence numbers
	actualSendPackets := s.App.RatelimitKeeper.GetAllPendingSendPackets(s.Ctx)
	s.Require().Equal(sendPackets, actualSendPackets, "all send packets")

	// Remove 0 sequence numbers and all sequence numbers from channel-0
	s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, "channel-0", 0)
	s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, "channel-1", 0)
	s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, "channel-0")

	// Check that only the remaining sequences are found
	for _, channelId := range []string{"channel-0", "channel-1"} {
		for sequence := uint64(0); sequence < 5; sequence++ {
			expected := (channelId == "channel-1") && (sequence != 0)
			actual := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().Equal(expected, actual, "send packet after removal - channel: %s, sequence: %d", channelId, sequence)
		}
	}
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

func (s *KeeperTestSuite) TestAddressWhitelist() {
	// Store addresses in whitelist
	expectedWhitelist := []types.WhitelistedAddressPair{
		{Sender: "sender-1", Receiver: "receiver-1"},
		{Sender: "sender-2", Receiver: "receiver-2"},
		{Sender: "sender-3", Receiver: "receiver-3"},
	}
	for _, addressPair := range expectedWhitelist {
		s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, addressPair)
	}

	// Confirm that each was found
	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, addressPair.Receiver)
		s.Require().True(found, "address pair should have been whitelisted (%s/%s)",
			addressPair.Sender, addressPair.Receiver)
	}

	// Confirm that looking both the sender and receiver must match for the pair to be whitelisted
	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, "fake-receiver")
		s.Require().False(found, "address pair should not have been whitelisted (%s/%s)",
			addressPair.Sender, "fake-receiver")

		found = s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, "fake-sender", addressPair.Receiver)
		s.Require().False(found, "address pair should not have been whitelisted (%s/%s)",
			"fake-sender", addressPair.Receiver)
	}

	// Check GetAll
	actualWhitelist := s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx)
	s.Require().Equal(expectedWhitelist, actualWhitelist, "whitelist get all")

	// Finally, remove each from whitelist
	for _, addressPair := range expectedWhitelist {
		s.App.RatelimitKeeper.RemoveWhitelistedAddressPair(s.Ctx, addressPair.Sender, addressPair.Receiver)
	}

	// Confirm there are no longer any whitelisted pairs
	actualWhitelist = s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx)
	s.Require().Empty(actualWhitelist, "whitelist should have been cleared")

	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, addressPair.Receiver)
		s.Require().False(found, "address pair should no longer be whitelisted (%s/%s)",
			addressPair.Sender, addressPair.Receiver)
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
	s.App.RatelimitKeeper.RemoveWhitelistedAddressPair(s.Ctx, sender, receiver)
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

		if action.addToWhitelist {
			s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, types.WhitelistedAddressPair{
				Sender:   sender,
				Receiver: receiver,
			})
			continue
		} else if action.removeFromWhitelist {
			s.App.RatelimitKeeper.RemoveWhitelistedAddressPair(s.Ctx, sender, receiver)
			continue
		}

		amount := sdkmath.NewInt(action.amount)
		packetInfo := keeper.RateLimitedPacketInfo{
			ChannelID: channelId,
			Denom:     denom,
			Amount:    amount,
			Sender:    sender,
			Receiver:  receiver,
		}
		updatedFlow, err := s.App.RatelimitKeeper.CheckRateLimitAndUpdateFlow(s.Ctx, action.direction, packetInfo)

		// Each action optionally errors or skips a flow update
		if action.expectedError != "" {
			s.Require().ErrorContains(err, action.expectedError, tc.name+" - action: #%d - error", i)
		} else {
			s.Require().NoError(err, tc.name+" - action: #%d - no error", i)

			expectedUpdateFlow := !action.skipFlowUpdate
			s.Require().Equal(expectedUpdateFlow, updatedFlow, tc.name+" - action: #%d - updated flow", i)

			if expectedUpdateFlow {
				if action.direction == types.PACKET_RECV {
					expectedInflow = expectedInflow.Add(amount)
				} else {
					expectedOutflow = expectedOutflow.Add(amount)
				}
			}
		}

		// Confirm flow is updated properly (or left as is if the theshold was exceeded)
		rateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found)
		s.Require().Equal(expectedInflow.Int64(), rateLimit.Flow.Inflow.Int64(), tc.name+" - action: #%d - inflow", i)
		s.Require().Equal(expectedOutflow.Int64(), rateLimit.Flow.Outflow.Int64(), tc.name+" - action: #%d - outflow", i)
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
				{direction: types.PACKET_SEND, amount: 6,
					expectedError: "Outflow exceeds quota"},
			},
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
				{direction: types.PACKET_RECV, amount: 6,
					expectedError: "Inflow exceeds quota"},
			},
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
				{direction: types.PACKET_RECV, amount: 6, //    +6, Net: +12 (exceeds threshold)
					expectedError: "Inflow exceeds quota"},
			},
		},
		{
			name: "send_then_recv_over_outflow",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6}, //   -6, Net: -6
				{direction: types.PACKET_RECV, amount: 2}, //   +2, Net: -4
				{direction: types.PACKET_SEND, amount: 6}, //   -6, Net: -10
				{direction: types.PACKET_RECV, amount: 2}, //   +2, Net: -8
				{direction: types.PACKET_SEND, amount: 6, //    -6, Net: -14 (exceeds threshold)
					expectedError: "Outflow exceeds quota"},
			},
		},
		{
			name: "recv_then_send_over_inflow",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +6
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +4
				{direction: types.PACKET_RECV, amount: 6}, //   +6, Net: +10
				{direction: types.PACKET_SEND, amount: 2}, //   -2, Net: +8
				{direction: types.PACKET_RECV, amount: 6, //    +6, Net: +14 (exceeds threshold)
					expectedError: "Inflow exceeds quota"},
			},
		},
		{
			name: "recv_then_send_over_outflow",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 2}, //  +2, Net: +2
				{direction: types.PACKET_SEND, amount: 6}, //  -6, Net: -4
				{direction: types.PACKET_RECV, amount: 2}, //  +2, Net: -2
				{direction: types.PACKET_SEND, amount: 6}, //  -6, Net: -8
				{direction: types.PACKET_RECV, amount: 2}, //  +2, Net: -6
				{direction: types.PACKET_SEND, amount: 10, //  +6, Net: -12 (exceeds threshold)
					expectedError: "Outflow exceeds quota"},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestCheckRateLimitAndUpdatedFlow_DenomBlacklist() {
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
				{direction: types.PACKET_SEND, amount: 6,
					expectedError: types.ErrDenomIsBlacklisted.Error()},
			},
		},
		{
			name: "send_recv_blacklist_recv",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_RECV, amount: 6,
					expectedError: types.ErrDenomIsBlacklisted.Error()},
			},
		},
		{
			name: "recv_send_blacklist_send",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_SEND, amount: 6,
					expectedError: types.ErrDenomIsBlacklisted.Error()},
			},
		},
		{
			name: "recv_send_blacklist_recv",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToBlacklist: true},
				{direction: types.PACKET_RECV, amount: 6,
					expectedError: types.ErrDenomIsBlacklisted.Error()},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestCheckRateLimitAndUpdatedFlow_AddressWhitelist() {
	testCases := []checkRateLimitTestCase{
		{
			name: "send_whitelist_send", // should succeed
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{addToWhitelist: true},
				{direction: types.PACKET_SEND, amount: 6, skipFlowUpdate: true},
			},
		},
		{
			name: "recv_whitelist_recv", // should succeed
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{addToWhitelist: true},
				{direction: types.PACKET_RECV, amount: 6, skipFlowUpdate: true},
			},
		},
		{
			name: "send_send_whitelist_send", // should succeed
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_SEND, amount: 6, expectedError: "Outflow exceeds quota"}, // fails
				{addToWhitelist: true},
				{direction: types.PACKET_SEND, amount: 6, skipFlowUpdate: true}, // succeeds
			},
		},
		{
			name: "recv_recv_whitelist_recv", // should succeed
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_RECV, amount: 6, expectedError: "Inflow exceeds quota"}, // fails
				{addToWhitelist: true},
				{direction: types.PACKET_RECV, amount: 6, skipFlowUpdate: true}, // succeeds
			},
		},
		{
			name: "send_recv_send_whitelist_send", // should succeed
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{addToWhitelist: true},
				{direction: types.PACKET_SEND, amount: 6, skipFlowUpdate: true},
			},
		},
		{
			name: "recv_send_recv_whitelist_recv", // should succeed
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{direction: types.PACKET_SEND, amount: 6},
				{direction: types.PACKET_RECV, amount: 6},
				{addToWhitelist: true},
				{direction: types.PACKET_RECV, amount: 6, skipFlowUpdate: true},
			},
		},
		{
			name: "add_then_remove_whitelist_recv",
			actions: []action{
				{direction: types.PACKET_RECV, amount: 6},
				{addToWhitelist: true},
				{removeFromWhitelist: true},
				{direction: types.PACKET_RECV, amount: 6, expectedError: "Inflow exceeds quota"},
			},
		},
		{
			name: "add_then_remove_whitelist_send",
			actions: []action{
				{direction: types.PACKET_SEND, amount: 6},
				{addToWhitelist: true},
				{removeFromWhitelist: true},
				{direction: types.PACKET_SEND, amount: 6, expectedError: "Outflow exceeds quota"},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.processCheckRateLimitAndUpdateFlowTestCase(tc)
		})
	}
}

func (s *KeeperTestSuite) TestUndoSendPacket() {
	// Helper function to check the rate limit outflow amount
	checkOutflow := func(channelId, denom string, expectedAmount sdkmath.Int) {
		rateLimit, found := s.App.RatelimitKeeper.GetRateLimit(s.Ctx, denom, channelId)
		s.Require().True(found, "rate limit should have been found")
		s.Require().Equal(expectedAmount.Int64(), rateLimit.Flow.Outflow.Int64(),
			"outflow - channel: %s, denom: %s", channelId, denom)
	}

	// Create two rate limits
	initialOutflow := sdkmath.NewInt(100)
	packetSendAmount := sdkmath.NewInt(10)
	rateLimit1 := types.RateLimit{
		Path: &types.Path{Denom: denom, ChannelId: channelId},
		Flow: &types.Flow{Outflow: initialOutflow},
	}
	rateLimit2 := types.RateLimit{
		Path: &types.Path{Denom: "different-denom", ChannelId: "different-channel"},
		Flow: &types.Flow{Outflow: initialOutflow},
	}
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit1)
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit2)

	// Store a pending packet sequence number of 2 for the first rate limit
	s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, 2)

	// Undo a send of 10 from the first rate limit, with sequence 1
	// If should NOT modify the outflow since sequence 1 was not sent in the current quota
	err := s.App.RatelimitKeeper.UndoSendPacket(s.Ctx, channelId, 1, denom, packetSendAmount)
	s.Require().NoError(err, "no error expected when undoing send packet sequence 1")

	checkOutflow(channelId, denom, initialOutflow)

	// Now undo a send from the same rate limit with sequence 2
	// If should decrement the outflow since 2 is in the current quota
	err = s.App.RatelimitKeeper.UndoSendPacket(s.Ctx, channelId, 2, denom, packetSendAmount)
	s.Require().NoError(err, "no error expected when undoing send packet sequence 2")

	checkOutflow(channelId, denom, initialOutflow.Sub(packetSendAmount))

	// Confirm the outflow of the second rate limit has not been touched
	checkOutflow("different-channel", "different-denom", initialOutflow)

	// Confirm sequence number was removed
	found := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, 2)
	s.Require().False(found, "packet sequence number should have been removed")
}
