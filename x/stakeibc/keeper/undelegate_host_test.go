package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const UndelegateHostZoneChainId = "evmos_9001-2" // the relevant zone for this test

func (s *KeeperTestSuite) SetupTestUndelegateHost(
	totalWeight int64,
	totalStake sdkmath.Int,
	unbondAmount sdkmath.Int,
	validators []*types.Validator,
) UnbondingTestCase {
	delegationAccountOwner := types.FormatICAAccountOwner(UndelegateHostZoneChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Sanity checks:
	//  - total stake matches
	//  - total weights sum to 100
	actualTotalStake := sdkmath.ZeroInt()
	actualTotalWeights := uint64(0)
	for _, validator := range validators {
		actualTotalStake = actualTotalStake.Add(validator.Delegation)
		actualTotalWeights += validator.Weight
	}
	s.Require().Equal(totalStake.Int64(), actualTotalStake.Int64(), "test setup failed - total stake does not match")
	s.Require().Equal(totalWeight, int64(actualTotalWeights), "test setup failed - total weight does not match")

	// Store the validators on the host zone
	hostZone := types.HostZone{
		ChainId:              UndelegateHostZoneChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		Validators:           validators,
		TotalDelegations:     totalStake,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store the total unbond amount across two epoch unbonding records
	halfUnbondAmount := unbondAmount.Quo(sdkmath.NewInt(2))
	for i := uint64(1); i <= 2; i++ {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: i,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        UndelegateHostZoneChainId,
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					NativeTokenAmount: halfUnbondAmount,
				},
			},
		})
	}

	// Mock the epoch tracker to timeout 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochstypes.DAY_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get tx seq number before the ICA was submitted to check whether an ICA was submitted
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found before ica")

	return UnbondingTestCase{
		hostZone:                   hostZone,
		totalUnbondAmount:          unbondAmount,
		delegationChannelID:        delegationChannelID,
		delegationPortID:           delegationPortID,
		channelStartSequence:       startSequence,
		expectedUnbondingRecordIds: []uint64{1, 2},
	}
}

// Helper function to check that an undelegation ICA was submitted and that the callback data
// holds the expected unbondings for each validator
func (s *KeeperTestSuite) CheckUndelegateHostMessages(tc UnbondingTestCase, expectedUnbondings []ValidatorUnbonding) {

	// Check that IsUndelegateHostPrevented(ctx) has not yet been flipped to true
	s.Require().False(s.App.StakeibcKeeper.IsUndelegateHostPrevented(s.Ctx))

	// Trigger unbonding
	err := s.App.StakeibcKeeper.UndelegateHostEvmos(s.Ctx, tc.totalUnbondAmount)
	s.Require().NoError(err, "no error expected when calling unbond from host")

	// Check that sequence number incremented from a sent ICA
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "sequence number should have incremented")

	// Check that callback data was stored
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbackData, 1, "there should only be one callback data stored")

	// Check host zone and epoch unbonding record id's
	var actualCallback types.UndelegateHostCallback
	err = proto.Unmarshal(callbackData[0].CallbackArgs, &actualCallback)
	s.Require().NoError(err, "no error expected when unmarshalling callback args")

	s.Require().Equal(tc.totalUnbondAmount, actualCallback.Amt, "amount on callback")

	// Check splits from callback data align with expected unbondings
	s.Require().Len(actualCallback.SplitDelegations, len(expectedUnbondings), "number of unbonding messages")
	for i, expected := range expectedUnbondings {
		actualSplit := actualCallback.SplitDelegations[i]
		s.Require().Equal(expected.Validator, actualSplit.Validator, "callback message validator - index %d", i)
		s.Require().Equal(expected.UnbondAmount.Int64(), actualSplit.Amount.Int64(), "callback message amount - index %d", i)
	}

	// Check the delegation change in progress was incremented from each that had an unbonding
	actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, UndelegateHostZoneChainId)
	s.Require().True(found, "host zone should have been found")

	for _, actualValidator := range actualHostZone.Validators {
		validatorUnbonded := false
		for _, unbondedVal := range expectedUnbondings {
			if actualValidator.Address == unbondedVal.Validator {
				validatorUnbonded = true
			}
		}

		expectedDelegationChangesInProgress := 0
		if validatorUnbonded {
			expectedDelegationChangesInProgress = 1
		}
		s.Require().Equal(expectedDelegationChangesInProgress, int(actualValidator.DelegationChangesInProgress),
			"validator %s delegation changes in progress", actualValidator.Address)
	}

	// Check that the unbond event was emitted with the proper unbond amount
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyTotalUnbondAmount, tc.totalUnbondAmount.String())
}

func (s *KeeperTestSuite) TestUndelegateHost_Successful_UnbondOnlyZeroWeightVals() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:        50
	// Stake After Unbond: 1200
	totalUnbondAmount := sdkmath.NewInt(50)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 100-120 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1200 = 420, Capacity: 420-420 = 0
		// No capacity -> unbondings
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1200 = 360, Capacity: 300-360 = -60 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1200 = 120, Capacity: 200 - 120 = 80
		// >>> Ratio: 110/200 = 0.55 -> #3 Priority <<<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1200 = 180, Capacity: 160-180 = -20 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 50 - 40 = 10 unbond remaining
		// valE has #2 priority - unbond up to remaining
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUndelegateHostMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUndelegateHost_Successful_UnbondTotalLessThanTotalLSM() {
	// Native Stake:       1000
	// LSM Stake:           250
	// Total Stake:        1250
	//
	// Unbond Amount:       150
	// Stake After Unbond: 1100
	totalUnbondAmount := sdkmath.NewInt(150)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 100-110 = -10 -> 0
		// No capacity -> unbondings
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 1100 = 385, Capacity: 420-385 = 35
		// >>> Ratio: 385/420 = 0.91 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 1100 = 330, Capacity: 300-330 = -30 -> 0
		// No capacity -> unbondings
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 1100 = 110, Capacity: 200 - 110 = 90
		// >>> Ratio: 110/200 = 0.55 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 1100 = 165, Capacity: 160-165 = -5 -> 0
		// No capacity -> unbondings
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 150 - 40 = 110 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 150 - 40 - 30 = 80 unbond remaining
		// valF has #3 priority - unbond up to remaining
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(80)},
	}

	tc := s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUndelegateHostMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUndelegateHost_Successful_UnbondTotalGreaterThanTotalLSM() {
	// Native Stake: 1000
	// LSM Stake:     250
	// Total Stake:  1250
	//
	// Unbond Amount:      350
	// Stake After Unbond: 900
	totalUnbondAmount := sdkmath.NewInt(350)
	totalStake := sdkmath.NewInt(1250)
	totalWeight := int64(100)

	validators := []*types.Validator{
		// Current: 100, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 100-90 = 10
		// >>> Ratio: 90/100 = 0.9 -> Priority #7 <<<
		{Address: "valA", Weight: 10, Delegation: sdkmath.NewInt(100)},
		// Current: 420, Weight: 35%, Balanced: 35% * 900 = 315, Capacity: 420-315 = 105
		// >>> Ratio: 315/420 = 0.75 -> Priority #4 <<<
		{Address: "valB", Weight: 35, Delegation: sdkmath.NewInt(420)},
		// Weight: 0%, Balanced: 0, Capacity: 40
		// >>> Ratio: 0 -> Priority #1 <<<
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(40)},
		// Current: 300, Weight: 30%, Balanced: 30% * 900 = 270, Capacity: 300-270 = 30
		// >>> Ratio: 270/300 = 0.9 -> Priority #6 <<<
		{Address: "valD", Weight: 30, Delegation: sdkmath.NewInt(300)},
		// Weight: 0%, Balanced: 0, Capacity: 30
		// >>> Ratio: 0 -> Priority #2 <<<
		{Address: "valE", Weight: 0, Delegation: sdkmath.NewInt(30)},
		// Current: 200, Weight: 10%, Balanced: 10% * 900 = 90, Capacity: 200 - 90 = 110
		// >>> Ratio: 90/200 = 0.45 -> Priority #3 <<<
		{Address: "valF", Weight: 10, Delegation: sdkmath.NewInt(200)},
		// Current: 160, Weight: 15%, Balanced: 15% * 900 = 135, Capacity: 160-135 = 25
		// >>> Ratio: 135/160 = 0.85 -> Priority #5 <<<
		{Address: "valG", Weight: 15, Delegation: sdkmath.NewInt(160)},
	}

	expectedUnbondings := []ValidatorUnbonding{
		// valC has #1 priority - unbond up to capacity at 40
		{Validator: "valC", UnbondAmount: sdkmath.NewInt(40)},
		// 350 - 40 = 310 unbond remaining
		// valE has #2 priority - unbond up to capacity at 30
		{Validator: "valE", UnbondAmount: sdkmath.NewInt(30)},
		// 310 - 30 = 280 unbond remaining
		// valF has #3 priority - unbond up to capacity at 110
		{Validator: "valF", UnbondAmount: sdkmath.NewInt(110)},
		// 280 - 110 = 170 unbond remaining
		// valB has #4 priority - unbond up to capacity at 105
		{Validator: "valB", UnbondAmount: sdkmath.NewInt(105)},
		// 170 - 105 = 65 unbond remaining
		// valG has #5 priority - unbond up to capacity at 25
		{Validator: "valG", UnbondAmount: sdkmath.NewInt(25)},
		// 65 - 25 = 40 unbond remaining
		// valD has #6 priority - unbond up to capacity at 30
		{Validator: "valD", UnbondAmount: sdkmath.NewInt(30)},
		// 40 - 30 = 10 unbond remaining
		// valA has #7 priority - unbond up to remaining
		{Validator: "valA", UnbondAmount: sdkmath.NewInt(10)},
	}

	tc := s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, validators)
	s.CheckUndelegateHostMessages(tc, expectedUnbondings)
}

func (s *KeeperTestSuite) TestUndelegateHost_AmountTooLarge() {
	// Call undelegateHost with an amount that is greater than the max amount, it should fail
	unbondAmount, ok := math.NewIntFromString("25000000000000000000000001")
	s.Require().True(ok, "could not parse unbondAmount")
	err := s.App.StakeibcKeeper.UndelegateHostEvmos(s.Ctx, unbondAmount)
	s.Require().ErrorContains(err, fmt.Sprintf("total unbond amount %v is greater than", unbondAmount))
}

func (s *KeeperTestSuite) TestUndelegateHost_ZeroUnbondAmount() {
	totalWeight := int64(0)
	totalStake := sdkmath.ZeroInt()
	totalUnbondAmount := sdkmath.ZeroInt()
	tc := s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, []*types.Validator{})

	// Call unbond - it should NOT error since the unbond amount was 0 - but it should short circuit
	err := s.App.StakeibcKeeper.UndelegateHostEvmos(s.Ctx, totalUnbondAmount)
	s.Require().Nil(err, "unbond should not have thrown an error - it should have simply ignored the host zone")

	// Confirm no ICAs were sent
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ica")
	s.Require().Equal(tc.channelStartSequence, endSequence, "sequence number should stay the same since no messages were sent")
}

func (s *KeeperTestSuite) TestUndelegateHost_ZeroValidatorWeights() {
	// Setup the test with all zero-weight validators
	totalWeight := int64(0)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(10)
	validators := []*types.Validator{
		{Address: "valA", Weight: 0, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 0, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 0, Delegation: sdkmath.NewInt(25)},
	}
	s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UndelegateHostEvmos(s.Ctx, totalUnbondAmount)
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestUndelegateHost_InsufficientDelegations() {
	// Setup the test where the total unbond amount is greater than the current delegations
	totalWeight := int64(100)
	totalStake := sdkmath.NewInt(100)
	totalUnbondAmount := sdkmath.NewInt(200)
	validators := []*types.Validator{
		{Address: "valA", Weight: 25, Delegation: sdkmath.NewInt(25)},
		{Address: "valB", Weight: 50, Delegation: sdkmath.NewInt(50)},
		{Address: "valC", Weight: 25, Delegation: sdkmath.NewInt(25)},
	}
	s.SetupTestUndelegateHost(totalWeight, totalStake, totalUnbondAmount, validators)

	// Call unbond - it should fail
	err := s.App.StakeibcKeeper.UndelegateHostEvmos(s.Ctx, totalUnbondAmount)
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero")
}
