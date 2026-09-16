package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"
	_ "github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// The fixture host zone has a 21 day unbonding period, so it unbonds every 4th day epoch
const (
	unbondingEpoch    = uint64(4)
	nonUnbondingEpoch = uint64(5)
)

type PendingUndelegationTestCase struct {
	hostZone            types.HostZone
	pendingAmount       sdkmath.Int
	delegationChannelID string
	delegationPortID    string
}

func (s *KeeperTestSuite) TestPendingUndelegation_SetGetRemove() {
	chainId := "chain-1"
	amount := sdkmath.NewInt(200_476_671)

	// Nothing should be found before the key is set
	_, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, chainId)
	s.Require().False(found, "pending undelegation should not be found before set")

	// Set and confirm the amount round-trips
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, chainId, amount)
	actualAmount, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, chainId)
	s.Require().True(found, "pending undelegation should be found after set")
	s.Require().Equal(amount, actualAmount, "pending undelegation amount")

	// Overwriting replaces the amount
	updatedAmount := sdkmath.NewInt(5)
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, chainId, updatedAmount)
	actualAmount, found = s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, chainId)
	s.Require().True(found, "pending undelegation should be found after overwrite")
	s.Require().Equal(updatedAmount, actualAmount, "pending undelegation amount after overwrite")

	// Remove and confirm it's gone
	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, chainId)
	_, found = s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, chainId)
	s.Require().False(found, "pending undelegation should not be found after remove")

	// Removing a missing key is a no-op
	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, chainId)
}

func (s *KeeperTestSuite) TestGetAllPendingUndelegations() {
	// With nothing stored, the list is empty
	s.Require().Empty(s.App.StakeibcKeeper.GetAllPendingUndelegations(s.Ctx), "no pending undelegations")

	// Store two entries and confirm both come back (ordered by chain id)
	expected := []types.PendingUndelegation{
		{ChainId: "chain-A", Amount: sdkmath.NewInt(100)},
		{ChainId: "chain-B", Amount: sdkmath.NewInt(200)},
	}
	for _, pending := range expected {
		s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, pending.ChainId, pending.Amount)
	}
	s.Require().Equal(expected, s.App.StakeibcKeeper.GetAllPendingUndelegations(s.Ctx), "all pending undelegations")

	// Removing one leaves the other
	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, "chain-A")
	s.Require().Equal(expected[1:], s.App.StakeibcKeeper.GetAllPendingUndelegations(s.Ctx), "pending undelegations after remove")
}

// Registers a host zone with a delegation ICA channel and two validators, and queues a pending
// undelegation. The validators are set up so that only val1 has unbond capacity for the pending amount:
//
//	Total Stake:  1000 (val1: 600, val2: 400), weights 50/50
//	Pending:       200 → balanced delegation after unbonding is 400/400
//	Capacity:      val1: 200, val2: 0
//
// The unbonding period is 21 days so the host zone only unbonds on every 4th day epoch
func (s *KeeperTestSuite) SetupSubmitPendingUndelegations() PendingUndelegationTestCase {
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	pendingAmount := sdkmath.NewInt(200)
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators: []*types.Validator{
			{Address: "val1", Weight: 50, Delegation: sdkmath.NewInt(600)},
			{Address: "val2", Weight: 50, Delegation: sdkmath.NewInt(400)},
		},
		TotalDelegations:    sdkmath.NewInt(1000),
		RedemptionRate:      sdkmath.LegacyOneDec(),
		MaxMessagesPerIcaTx: 32,
		UnbondingPeriod:     21,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, pendingAmount)
	s.Require().Equal(unbondingEpoch, hostZone.GetUnbondingFrequency(), "fixture unbonding frequency")

	// Mock the day epoch tracker so the ICA timeout can be computed
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	})

	return PendingUndelegationTestCase{
		hostZone:            hostZone,
		pendingAmount:       pendingAmount,
		delegationChannelID: delegationChannelID,
		delegationPortID:    delegationPortID,
	}
}

// Asserts the pending key is still stored with the original amount, and that no validator was
// flagged with a delegation change in progress
func (s *KeeperTestSuite) checkPendingUndelegationNotSubmitted(tc PendingUndelegationTestCase) {
	actualAmount, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation should be kept for retry")
	s.Require().Equal(tc.pendingAmount, actualAmount, "pending undelegation amount should be unchanged")

	hostZone := s.MustGetHostZone(HostChainId)
	for _, validator := range hostZone.Validators {
		s.Require().Zero(validator.DelegationChangesInProgress, "validator %s delegation changes in progress", validator.Address)
	}
	s.Require().Empty(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), "no callback data should be stored")
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "nothing should be marked in flight")
}

func (s *KeeperTestSuite) TestSubmitPendingUndelegations_Successful() {
	tc := s.SetupSubmitPendingUndelegations()

	// Submit and confirm exactly one ICA was sent
	s.CheckICATxSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})

	// The amount stays pending until the batch acks; the in-flight count blocks a resubmission
	actualAmount, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation should stay stored while the batch is in flight")
	s.Require().Equal(tc.pendingAmount, actualAmount, "pending amount unchanged at submission")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "one batch in flight")

	// The callback should carry the split for val1 only, with no epoch unbonding record ids
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1, "one callback data should be stored")

	var callback types.UndelegateCallback
	err := proto.Unmarshal(callbackData[0].CallbackArgs, &callback)
	s.Require().NoError(err, "no error expected when unmarshalling callback args")
	s.Require().Equal(HostChainId, callback.HostZoneId, "callback host zone")
	s.Require().Nil(callback.EpochUnbondingRecordIds, "callback should have no epoch unbonding record ids")
	s.Require().Len(callback.SplitUndelegations, 1, "callback should have one split")
	s.Require().Equal("val1", callback.SplitUndelegations[0].Validator, "callback split validator")
	s.Require().Equal(tc.pendingAmount, callback.SplitUndelegations[0].NativeTokenAmount, "callback split amount")

	// Only val1 should have a delegation change in progress
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(1), hostZone.Validators[0].DelegationChangesInProgress, "val1 delegation changes in progress")
	s.Require().Zero(hostZone.Validators[1].DelegationChangesInProgress, "val2 delegation changes in progress")

	// The delegation balances are untouched until the callback
	s.Require().Equal(tc.hostZone.TotalDelegations, hostZone.TotalDelegations, "total delegations unchanged before callback")

	// The undelegation event should have been emitted with the pending amount
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, tc.pendingAmount.String())

	// A second call should find nothing pending and submit nothing
	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})
}

// On a host zone's unbonding epoch the normal unbonding flow already targets the validators' capacity,
// so the pending undelegation must be deferred to the next day epoch rather than submitted alongside it
func (s *KeeperTestSuite) TestSubmitPendingUndelegations_UnbondingEpochSkipped() {
	tc := s.SetupSubmitPendingUndelegations()

	// On the unbonding epoch nothing should be submitted and the key should be kept
	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, unbondingEpoch)
		return nil
	})
	s.checkPendingUndelegationNotSubmitted(tc)

	// On the following epoch it should be submitted, with the amount kept and the batch in flight
	s.CheckICATxSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, unbondingEpoch+1)
		return nil
	})
	_, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation stays stored after submission on a non-unbonding epoch")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "one batch in flight")
}

func (s *KeeperTestSuite) TestSubmitPendingUndelegations_NothingPending() {
	tc := s.SetupSubmitPendingUndelegations()
	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, HostChainId)

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})

	hostZone := s.MustGetHostZone(HostChainId)
	for _, validator := range hostZone.Validators {
		s.Require().Zero(validator.DelegationChangesInProgress, "validator %s delegation changes in progress", validator.Address)
	}
}

func (s *KeeperTestSuite) TestSubmitPendingUndelegations_ChannelClosed() {
	tc := s.SetupSubmitPendingUndelegations()

	// Close the delegation channel so the ICA submission fails
	s.UpdateChannelState(tc.delegationPortID, tc.delegationChannelID, channeltypes.CLOSED)

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})
	s.checkPendingUndelegationNotSubmitted(tc)
}

func (s *KeeperTestSuite) TestSubmitPendingUndelegations_ZeroWeights() {
	tc := s.SetupSubmitPendingUndelegations()

	// Zero out the validator weights so no balanced delegation can be computed
	hostZone := tc.hostZone
	for _, validator := range hostZone.Validators {
		validator.Weight = 0
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})
	s.checkPendingUndelegationNotSubmitted(tc)
}

// A validator that's fully drained (zero weight) and slashed (SharesToTokensRate < 1) has a tiny
// rounding buffer withheld from its undelegation (see applySharesRoundingSafety in unbonding.go).
// With no other validator left with spare capacity to absorb that buffer, the cascade falls short
// of the full pending amount and GetUnbondingICAMessages returns "unable to unbond full amount" -
// a real insufficient-capacity failure, distinct from the zero-weight GetTargetValAmtsForHostZone
// error covered by TestSubmitPendingUndelegations_ZeroWeights
func (s *KeeperTestSuite) TestSubmitPendingUndelegations_InsufficientCapacity() {
	tc := s.SetupSubmitPendingUndelegations()

	hostZone := tc.hostZone
	hostZone.Validators = []*types.Validator{
		{
			Address:            "val1",
			Weight:             0,
			Delegation:         sdkmath.NewInt(100),
			SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.99"),
		},
		{
			Address:            "val2",
			Weight:             100,
			Delegation:         sdkmath.NewInt(900),
			SharesToTokensRate: sdkmath.LegacyOneDec(),
		},
	}
	hostZone.TotalDelegations = sdkmath.NewInt(1000)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(100))

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})

	tc.pendingAmount = sdkmath.NewInt(100)
	s.checkPendingUndelegationNotSubmitted(tc)
}

// A pending amount of zero has no undelegate messages to submit. Submitting an ICA with no
// messages (or emitting an undelegation event) would be a no-op that clears the key for nothing,
// so the key must be kept and no ICA sent
func (s *KeeperTestSuite) TestSubmitPendingUndelegations_ZeroAmount() {
	tc := s.SetupSubmitPendingUndelegations()

	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, sdkmath.ZeroInt())

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})

	tc.pendingAmount = sdkmath.ZeroInt()
	s.checkPendingUndelegationNotSubmitted(tc)
}

func (s *KeeperTestSuite) TestSubmitPendingUndelegations_HostZoneNotFound() {
	tc := s.SetupSubmitPendingUndelegations()

	// Queue a pending amount for a chain with no host zone, alongside the valid one
	missingChainId := "missing-chain"
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, missingChainId, sdkmath.NewInt(50))

	// Only the valid host zone's ICA should be submitted
	s.CheckICATxSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})

	// The missing one is dropped; the valid one is submitted and stays pending with a batch in flight
	_, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, missingChainId)
	s.Require().False(found, "pending undelegation for a missing host zone should be removed")
	_, found = s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation for the valid host zone stays stored while in flight")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "one batch in flight")
	s.Require().Len(s.App.StakeibcKeeper.GetAllPendingUndelegations(s.Ctx), 1, "only the valid pending undelegation remains")
}

// A batch that is still awaiting its ack must block a second submission of the same amount
func (s *KeeperTestSuite) TestSubmitPendingUndelegations_InFlightSkipped() {
	tc := s.SetupSubmitPendingUndelegations()
	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, HostChainId, 1)

	s.CheckICATxNotSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})
	actualAmount, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation kept while a batch is in flight")
	s.Require().Equal(tc.pendingAmount, actualAmount, "pending amount untouched")
	s.Require().Empty(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), "no new callback data")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "in-flight count untouched")

	// Once the batch is released the next epoch submits
	s.App.StakeibcKeeper.RemovePendingUndelegationInFlight(s.Ctx, HostChainId)
	s.CheckICATxSubmitted(tc.delegationPortID, tc.delegationChannelID, func() error {
		s.App.StakeibcKeeper.SubmitPendingUndelegations(s.Ctx, nonUnbondingEpoch)
		return nil
	})
}

func (s *KeeperTestSuite) TestPendingUndelegationInFlight_Counter() {
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "nothing in flight initially")

	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, HostChainId, 2)
	s.Require().Equal(uint64(2), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))

	s.App.StakeibcKeeper.DecrementPendingUndelegationInFlight(s.Ctx, HostChainId)
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))

	s.App.StakeibcKeeper.DecrementPendingUndelegationInFlight(s.Ctx, HostChainId)
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "last decrement clears the key")

	// Decrementing an absent counter is a no-op, and the counters are per chain
	s.App.StakeibcKeeper.DecrementPendingUndelegationInFlight(s.Ctx, HostChainId)
	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, "other-chain", 3)
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))
	s.Require().Equal(uint64(3), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, "other-chain"))

	s.App.StakeibcKeeper.RemovePendingUndelegationInFlight(s.Ctx, "other-chain")
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, "other-chain"))
}

func (s *KeeperTestSuite) TestConsumePendingUndelegation() {
	// Consuming when nothing is pending is a no-op
	s.App.StakeibcKeeper.ConsumePendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(10))
	_, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().False(found)

	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(100))

	// A partial batch leaves the remainder pending
	s.App.StakeibcKeeper.ConsumePendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(30))
	remaining, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(sdkmath.NewInt(70), remaining)

	// The last batch clears the key (an over-consume clears it too rather than going negative)
	s.App.StakeibcKeeper.ConsumePendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(80))
	_, found = s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().False(found, "pending undelegation cleared once fully undelegated")
}
