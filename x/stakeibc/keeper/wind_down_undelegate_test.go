// x/stakeibc/keeper/wind_down_undelegate_test.go
package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochstypes "github.com/Stride-Labs/stride/v34/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type undelegateFromValidatorsTestCase struct {
	hostZone            types.HostZone
	delegationPortID    string
	delegationChannelID string
}

// Four validators: two unslashed, one slashed (rate < 1, so a full drain gets the rounding
// buffer), one with no delegation (skipped by the empty-list drain). Batch size 2 so three
// messages take two ICAs.
func (s *KeeperTestSuite) SetupUndelegateFromValidators() undelegateFromValidatorsTestCase {
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	validators := []*types.Validator{
		{Address: "val1", Delegation: sdkmath.NewInt(1000), SharesToTokensRate: sdkmath.LegacyOneDec()},
		{Address: "val2", Delegation: sdkmath.NewInt(2000), SharesToTokensRate: sdkmath.LegacyOneDec()},
		{Address: "val3", Delegation: sdkmath.NewInt(3000), SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.9")},
		{Address: "val4", Delegation: sdkmath.ZeroInt(), SharesToTokensRate: sdkmath.LegacyOneDec()},
	}
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		MaxMessagesPerIcaTx:  2,
		TotalDelegations:     sdkmath.NewInt(6000),
		Validators:           validators,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// The undelegate ICA timeout is read from the day epoch tracker
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000),
	})

	return undelegateFromValidatorsTestCase{
		hostZone:            hostZone,
		delegationPortID:    delegationPortID,
		delegationChannelID: delegationChannelID,
	}
}

func (s *KeeperTestSuite) setHostZoneUnbondingStatus(status recordtypes.HostZoneUnbonding_Status, amount int64) {
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{{
			HostZoneId:        HostChainId,
			Status:            status,
			NativeTokenAmount: sdkmath.NewInt(amount),
			StTokenAmount:     sdkmath.NewInt(amount),
		}},
	})
}

// The recorded delegations must never move: the callback does that on ack
func (s *KeeperTestSuite) checkNoAccountingMutation(tc undelegateFromValidatorsTestCase) {
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(tc.hostZone.TotalDelegations, hostZone.TotalDelegations, "total delegations unchanged")
	for i, validator := range hostZone.Validators {
		s.Require().Equal(tc.hostZone.Validators[i].Delegation, validator.Delegation, "%s delegation unchanged", validator.Address)
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_EmptyListDrainsEveryFundedValidator() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	numBatches, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), numBatches, "3 messages in batches of 2")

	endSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)
	s.Require().Equal(startSequence+2, endSequence, "two ICAs submitted")

	// val1..3 are flagged, val4 (no delegation) is untouched
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	// DelegationChangesInProgress is int64 (validator.pb.go); testify's Equal is type-strict
	s.Require().Equal(int64(1), hostZone.Validators[0].DelegationChangesInProgress)
	s.Require().Equal(int64(1), hostZone.Validators[1].DelegationChangesInProgress)
	s.Require().Equal(int64(1), hostZone.Validators[2].DelegationChangesInProgress)
	s.Require().Equal(int64(0), hostZone.Validators[3].DelegationChangesInProgress)

	// Both batches are registered in flight so the record-less callback's decrement is clean
	s.Require().Equal(uint64(2), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))

	// The callbacks carry no epoch unbonding record ids and the exact splits
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 2)
	splitsByValidator := map[string]sdkmath.Int{}
	for _, data := range callbackData {
		var callback types.UndelegateCallback
		s.Require().NoError(proto.Unmarshal(data.CallbackArgs, &callback))
		s.Require().Empty(callback.EpochUnbondingRecordIds, "record-less undelegation")
		s.Require().Equal(HostChainId, callback.HostZoneId)
		for _, split := range callback.SplitUndelegations {
			splitsByValidator[split.Validator] = split.NativeTokenAmount
		}
	}
	s.Require().Equal(sdkmath.NewInt(1000), splitsByValidator["val1"])
	s.Require().Equal(sdkmath.NewInt(2000), splitsByValidator["val2"])
	s.Require().Equal(sdkmath.NewInt(2999), splitsByValidator["val3"], "slashed validator drained with a 1 base unit buffer")
	s.Require().NotContains(splitsByValidator, "val4")

	s.checkNoAccountingMutation(tc)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, "5999")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_ExplicitListWithOffset() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{
		{Address: "val2", Offset: sdkmath.NewInt(500)},
	})
	numBatches, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), numBatches)
	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID))

	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().Equal(int64(0), hostZone.Validators[0].DelegationChangesInProgress, "val1 not listed")
	s.Require().Equal(int64(1), hostZone.Validators[1].DelegationChangesInProgress, "val2 listed")

	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1)
	var callback types.UndelegateCallback
	s.Require().NoError(proto.Unmarshal(callbackData[0].CallbackArgs, &callback))
	s.Require().Len(callback.SplitUndelegations, 1)
	s.Require().Equal(sdkmath.NewInt(1500), callback.SplitUndelegations[0].NativeTokenAmount, "2000 - 500 offset")
	s.Require().Equal(uint64(1), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))
	s.checkNoAccountingMutation(tc)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_InFlightAccumulates() {
	s.SetupUndelegateFromValidators()
	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, HostChainId, 3)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "added to, not overwritten")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_HaltedZoneAccepted() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err, "the drain does not depend on the zone being active")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsFlaggedValidator() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Validators[1].DelegationChangesInProgress = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}, {Address: "val2"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrInvalidDelegationsInProgress)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")
	s.Require().Equal(uint64(0), s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId))
}

// A validator mid-slash-query is excluded, as in the record-driven path; ops wait for the
// day-0 refresh callbacks (spec §9 step 1) before draining
func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsSlashQueryInProgress() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Validators[2].SlashQueryInProgress = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	// explicitly listed
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val3"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "slash query in progress")

	// and swept up by the empty list
	_, err = s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, types.NewMsgUndelegateFromValidators("admin", HostChainId, nil))
	s.Require().ErrorContains(err, "slash query in progress")
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")
}

// A stored v34-style pending undelegation would be consumed by the drain's record-less ack
func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsPendingUndelegation() {
	tc := s.SetupUndelegateFromValidators()
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(100))
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, types.NewMsgUndelegateFromValidators("admin", HostChainId, nil))
	s.Require().ErrorContains(err, "pending undelegation")
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")

	s.App.StakeibcKeeper.RemovePendingUndelegation(s.Ctx, HostChainId)
	_, err = s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, types.NewMsgUndelegateFromValidators("admin", HostChainId, nil))
	s.Require().NoError(err, "accepted once the pending amount is gone")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsDeprecatedZone() {
	tc := s.SetupUndelegateFromValidators()
	tc.hostZone.Deprecated = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "deprecated")
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsUnknownZone() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", "unknown-1", nil)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsUnknownValidator() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val9"}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrValidatorNotFound)
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsNonPositiveAmount() {
	s.SetupUndelegateFromValidators()

	// offset equal to the delegation
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1", Offset: sdkmath.NewInt(1000)}})
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "not positive")

	// validator with no delegation, explicitly listed
	msg = types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val4"}})
	_, err = s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorContains(err, "not positive")
}

// STRIDE-07 guard: a record still queued or retrying could never be submitted once the zone
// is drained, so the drain refuses until the day epoch has moved it to IN_PROGRESS
func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsQueuedRecords() {
	tc := s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, nil)

	for _, status := range []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
	} {
		s.setHostZoneUnbondingStatus(status, 100)
		startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)
		_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
		s.Require().ErrorIs(err, types.ErrHostZoneUnbondingPending, "status %s", status)
		s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")
	}
}

func (s *KeeperTestSuite) TestUndelegateFromValidators_AcceptsRecordsPastTheQueue() {
	s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}})

	for _, status := range []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		recordtypes.HostZoneUnbonding_CLAIMABLE,
	} {
		s.setHostZoneUnbondingStatus(status, 100)
		// clear the flag the previous iteration set so the validator is eligible again
		hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
		hostZone.Validators[0].DelegationChangesInProgress = 0
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

		_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
		s.Require().NoError(err, "status %s must not block the drain", status)
	}

	// A queued record with a zero amount is not a real record
	s.setHostZoneUnbondingStatus(recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, 0)
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators[0].DelegationChangesInProgress = 0
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err, "zero-amount queued record does not block")
}

// Pure builder: message shape and rounding safety, including a large slashed delegation
// where the 1e17 divisor yields a non-trivial buffer
func (s *KeeperTestSuite) TestBuildUndelegateFromValidatorsMsgs() {
	big := sdkmath.NewInt(5).Mul(sdkmath.NewInt(1e17)) // 5e17
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators: []*types.Validator{
			{Address: "val1", Delegation: sdkmath.NewInt(1000), SharesToTokensRate: sdkmath.LegacyOneDec()},
			{Address: "val2", Delegation: big, SharesToTokensRate: sdkmath.LegacyMustNewDecFromStr("0.95")},
		},
	}

	msgs, splits, err := s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(hostZone, nil)
	s.Require().NoError(err)
	s.Require().Len(msgs, 2)
	s.Require().Len(splits, 2)

	undelegate1 := msgs[0].(*stakingtypes.MsgUndelegate)
	s.Require().Equal("cosmos_DELEGATION", undelegate1.DelegatorAddress)
	s.Require().Equal("val1", undelegate1.ValidatorAddress)
	s.Require().Equal(Atom, undelegate1.Amount.Denom)
	s.Require().Equal(sdkmath.NewInt(1000), undelegate1.Amount.Amount, "unslashed full drain has no buffer")

	undelegate2 := msgs[1].(*stakingtypes.MsgUndelegate)
	s.Require().Equal(big.Sub(sdkmath.NewInt(5)), undelegate2.Amount.Amount, "5e17 / 1e17 = 5 base unit buffer on a slashed full drain")
	s.Require().Equal(undelegate2.Amount.Amount, splits[1].NativeTokenAmount)

	// A partial drain of the slashed validator gets no buffer
	msgs, _, err = s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(hostZone, []types.ValidatorUndelegation{{Address: "val2", Offset: sdkmath.NewInt(1)}})
	s.Require().NoError(err)
	s.Require().Equal(big.Sub(sdkmath.NewInt(1)), msgs[0].(*stakingtypes.MsgUndelegate).Amount.Amount)

	// Nothing to drain
	_, _, err = s.App.StakeibcKeeper.BuildUndelegateFromValidatorsMsgs(types.HostZone{ChainId: HostChainId}, nil)
	s.Require().ErrorContains(err, "no validator")
}

func (s *KeeperTestSuite) TestMsgServer_UndelegateFromValidators() {
	tc := s.SetupUndelegateFromValidators()
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)

	resp, err := s.GetMsgServer().UndelegateFromValidators(s.Ctx, types.NewMsgUndelegateFromValidators("admin", HostChainId, nil))
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), resp.NumBatchesSubmitted)
	s.Require().Equal(startSequence+2, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID))
}

// Spec section 6: a record-driven batch whose undelegate ack has not landed blocks the drain,
// whatever the record status; once acked (counter back to zero) an UNBONDING_IN_PROGRESS
// record that is only waiting out the unbonding period does not
func (s *KeeperTestSuite) TestUndelegateFromValidators_RejectsRecordBatchesInFlight() {
	tc := s.SetupUndelegateFromValidators()
	msg := types.NewMsgUndelegateFromValidators("admin", HostChainId, []types.ValidatorUndelegation{{Address: "val1"}})

	s.setHostZoneUnbondingStatus(recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS, 100)
	s.setUndelegationTxsInProgress(2)
	startSequence := s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID)
	_, err := s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrHostZoneUnbondingPending)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(tc.delegationPortID, tc.delegationChannelID), "nothing submitted")

	s.setUndelegationTxsInProgress(0)
	_, err = s.App.StakeibcKeeper.UndelegateFromValidators(s.Ctx, msg)
	s.Require().NoError(err, "acked in-progress record must not block")
}

func (s *KeeperTestSuite) setUndelegationTxsInProgress(count uint64) {
	record, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 1)
	s.Require().True(found)
	record.HostZoneUnbondings[0].UndelegationTxsInProgress = count
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, record)
}
