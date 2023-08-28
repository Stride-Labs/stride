package keeper_test

import (
	"fmt"
	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type RebalanceDelegationsForHostZoneTestCase struct {
	expectedRebalancings []types.Rebalancing
	channelStartSequence uint64
	hostZone             types.HostZone
	delegationChannelID  string
	delegationPortID     string
}

func (s *KeeperTestSuite) SetupTestRebalanceDelegationsForHostZone() RebalanceDelegationsForHostZoneTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Add host zone and validators
	delegationAddress := "cosmos_DELEGATION"
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		HostDenom:            Atom,
		DelegationIcaAddress: delegationAddress,
		ConnectionId:         ibctesting.FirstConnectionID,
		Validators: []*types.Validator{
			// Total delegation: 10000
			{Address: "val1", Weight: 25, Delegation: sdkmath.NewInt(3500)}, // Expected: 2500
			{Address: "val2", Weight: 50, Delegation: sdkmath.NewInt(2000)}, // Expected: 5000
			{Address: "val3", Weight: 25, Delegation: sdkmath.NewInt(4500)}, // Expected: 2500
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Add the stride epoch to determine the ICA timeout
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Build expected redelegation messages
	expectedRebalancings := []types.Rebalancing{
		{SrcValidator: "val3", DstValidator: "val2", Amt: sdkmath.NewInt(2000)}, // 2000 from val3 to val2
		{SrcValidator: "val1", DstValidator: "val2", Amt: sdkmath.NewInt(1000)}, // 1000 from val1 to val2
	}

	// Get the next sequence number to confirm if an ICA was sent
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found before ICA")

	return RebalanceDelegationsForHostZoneTestCase{
		expectedRebalancings: expectedRebalancings,
		channelStartSequence: startSequence,
		hostZone:             hostZone,
		delegationChannelID:  delegationChannelID,
		delegationPortID:     delegationPortID,
	}
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_Successful() {
	tc := s.SetupTestRebalanceDelegationsForHostZone()

	// Call rebalance
	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, HostChainId)
	s.Require().NoError(err, "no error expected with successful rebalancing")

	// Check that the ICA was sent by confirming the sequence number incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ICA")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "sequence number should have been incremented from ICA submission")

	// Check callback data
	allCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(allCallbackData, 1, "length of callback data")

	var callbackData types.RebalanceCallback
	err = proto.Unmarshal(allCallbackData[0].CallbackArgs, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")
	s.Require().Equal(HostChainId, callbackData.HostZoneId, "callback data chain-id")

	// Check splits from callback data
	actualRebalancings := callbackData.Rebalancings
	s.Require().Len(actualRebalancings, len(tc.expectedRebalancings), "number of rebalancings from callback data")

	expectedDelegationChanges := map[string]int{}
	for i, expected := range tc.expectedRebalancings {
		actual := actualRebalancings[i]
		s.Require().Equal(expected.SrcValidator, actual.SrcValidator, "rebalancing %d source validator")
		s.Require().Equal(expected.DstValidator, actual.DstValidator, "rebalancing %d destination validator")
		s.Require().Equal(expected.Amt.Int64(), actual.Amt.Int64(), "rebalancing %d amount")

		// Store the number of expected delegation changes for each validator
		if _, ok := expectedDelegationChanges[expected.SrcValidator]; !ok {
			expectedDelegationChanges[expected.SrcValidator] = 0
		}
		if _, ok := expectedDelegationChanges[expected.SrcValidator]; !ok {
			expectedDelegationChanges[expected.DstValidator] = 0
		}
		expectedDelegationChanges[expected.SrcValidator] += 1
		expectedDelegationChanges[expected.DstValidator] += 1
	}

	// Check the delegation change in progress was incremented from each redelegation
	actualHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")

	for _, actualValidator := range actualHostZone.Validators {
		expectedDelegationChangesInProgress := expectedDelegationChanges[actualValidator.Address]
		s.Require().Equal(expectedDelegationChangesInProgress, int(actualValidator.DelegationChangesInProgress),
			"validator %s delegation changes in progress", actualValidator.Address)
	}
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_SuccessfulBatchSend() {
	tc := s.SetupTestRebalanceDelegationsForHostZone()

	// Create 5 batches of redelegation messages
	// For each batch create the RebalanceIcaBatchSize number of validator pairs
	//  where the rebalance is going from one validator to the next
	// This will result in 5 ICA messages submitted
	numBatches := 5
	validators := []*types.Validator{}
	for batch := 1; batch <= numBatches; batch++ {
		for msg := 1; msg <= keeper.RebalanceIcaBatchSize; msg++ {
			validators = append(validators, []*types.Validator{
				{Address: fmt.Sprintf("src_val_%d_%d", batch, msg), Weight: 1, Delegation: sdkmath.NewInt(2)},
				{Address: fmt.Sprintf("dst_val_%d_%d", batch, msg), Weight: 1, Delegation: sdkmath.NewInt(0)},
			}...)
		}
	}
	hostZone := tc.hostZone
	hostZone.Validators = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Call rebalance
	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, HostChainId)
	s.Require().NoError(err, "no error expected with successful rebalancing")

	// Check that the ICA was sent by confirming the sequence number incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after ICA")
	s.Require().Equal(int(tc.channelStartSequence)+numBatches, int(endSequence),
		"sequence number should have been incremented multiple times from ICA submissions")
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_HostNotFound() {
	s.SetupTestRebalanceDelegationsForHostZone()

	// Attempt to rebalance with a host zone that does not exist - it should error
	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, "fake_host_zone")
	s.Require().ErrorContains(err, "Host zone fake_host_zone not found")
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_MissingDelegationAddress() {
	tc := s.SetupTestRebalanceDelegationsForHostZone()

	// Remove the delegation address from the host and then call rebalance - it should fail
	invalidHostZone := tc.hostZone
	invalidHostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, HostChainId)
	s.Require().ErrorContains(err, "no delegation account found for GAIA")
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_ZeroWeightValidators() {
	tc := s.SetupTestRebalanceDelegationsForHostZone()

	// Update the host zone validators so there are only 0 weight validators - rebalance should fail
	invalidHostZone := tc.hostZone
	invalidHostZone.Validators = []*types.Validator{{Address: "val1", Weight: 0}}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, HostChainId)
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero (0)")
}

func (s *KeeperTestSuite) TestRebalanceDelegationsForHostZone_FailedToSubmitICA() {
	tc := s.SetupTestRebalanceDelegationsForHostZone()

	// Remove the connection ID from the host zone so the ICA fails
	invalidHostZone := tc.hostZone
	invalidHostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	err := s.App.StakeibcKeeper.RebalanceDelegationsForHostZone(s.Ctx, HostChainId)
	s.Require().ErrorContains(err, "Failed to SubmitTxs for GAIA")
}

// Given a set of validator deltas (containing the expected change in delegation for each validator)
// and a set of expected rebalancings (containing the individual rebalance messages), calls
// RebalanceICAMessages and checks that the corresponding ICA messages match the expected rebalancings
func (s *KeeperTestSuite) checkRebalanceICAMessages(
	validatorDeltas []keeper.RebalanceValidatorDelegationChange,
	expectedRebalancings []types.Rebalancing,
) {
	// Build the expected ICA messages from the list of rebalancings above
	delegationAddress := "cosmos_DELEGATION"
	expectedMsgs := []proto.Message{}
	for _, rebalancing := range expectedRebalancings {
		expectedMsgs = append(expectedMsgs, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    delegationAddress,
			ValidatorSrcAddress: rebalancing.SrcValidator,
			ValidatorDstAddress: rebalancing.DstValidator,
			Amount:              sdk.NewCoin(Atom, rebalancing.Amt),
		})
	}

	// Only the validator address is needed in the host zone validator array
	hostZone := types.HostZone{
		HostDenom:            Atom,
		DelegationIcaAddress: delegationAddress, // used as ICA message sender
	}

	// Shuffle the validatorDeltas to ensure the sorting worked
	rand.Shuffle(len(validatorDeltas), func(i, j int) {
		validatorDeltas[i], validatorDeltas[j] = validatorDeltas[j], validatorDeltas[i]
	})

	// Get the rebalancing messages
	actualMsgs, actualRebalancings := s.App.StakeibcKeeper.GetRebalanceICAMessages(hostZone, validatorDeltas)

	// Confirm the rebalancing list used for the callback
	s.Require().Len(actualRebalancings, len(expectedRebalancings), "length of rebalancings")
	for i, expected := range expectedRebalancings {
		s.Require().Equal(expected.SrcValidator, actualRebalancings[i].SrcValidator, "rebalancing src validator, index %d", i)
		s.Require().Equal(expected.DstValidator, actualRebalancings[i].DstValidator, "rebalancing dst validator, index %d", i)
		s.Require().Equal(expected.Amt.Int64(), actualRebalancings[i].Amt.Int64(),
			"rebalancing amount, src: %s, dst: %s, index: %d", expected.SrcValidator, expected.DstValidator, i)
	}

	// Confirm the ICA messages list
	s.Require().Len(actualMsgs, len(expectedMsgs), "length of messages")
	for i, expectedMsg := range expectedMsgs {
		actual := actualMsgs[i].(*stakingtypes.MsgBeginRedelegate)
		expected := expectedMsg.(*stakingtypes.MsgBeginRedelegate)
		s.Require().Equal(delegationAddress, actual.DelegatorAddress, "message delegator address, index %d", i)
		s.Require().Equal(expected.ValidatorSrcAddress, actual.ValidatorSrcAddress, "message src validator, index %d", i)
		s.Require().Equal(expected.ValidatorDstAddress, actual.ValidatorDstAddress, "message dst validator, index %d", i)
	}
}

func (s *KeeperTestSuite) TestGetRebalanceICAMessages_EvenNumberValidators() {
	// Build up deltas for each validator, i.e. how much each validator needs to change by
	validatorDeltas := []keeper.RebalanceValidatorDelegationChange{
		// Overweight validators - they should lose some of their stake
		{ValidatorAddress: "val1", Delta: sdkmath.NewInt(21)}, // 15 to val10, 6 to val9
		{ValidatorAddress: "val2", Delta: sdkmath.NewInt(19)}, // 5 to val9, 11 to val8, 3 to val7
		{ValidatorAddress: "val3", Delta: sdkmath.NewInt(13)}, // 3 to val7, 5 to val6, 4 to val5, 1 to val4

		// Underweight validators - they should gain stake
		{ValidatorAddress: "val4", Delta: sdkmath.NewInt(-1)},   // 1 from val3
		{ValidatorAddress: "val5", Delta: sdkmath.NewInt(-4)},   // 4 from val3
		{ValidatorAddress: "val6", Delta: sdkmath.NewInt(-5)},   // 5 from val3
		{ValidatorAddress: "val7", Delta: sdkmath.NewInt(-6)},   // 3 from val2, 3 from val3
		{ValidatorAddress: "val8", Delta: sdkmath.NewInt(-11)},  // 11 from val2
		{ValidatorAddress: "val9", Delta: sdkmath.NewInt(-11)},  // 6 from val1, 5 from val2
		{ValidatorAddress: "val10", Delta: sdkmath.NewInt(-15)}, // 15 from val1
	}

	// Build up the expected messages, moving across the list above
	expectedRebalancings := []types.Rebalancing{
		{SrcValidator: "val1", DstValidator: "val10", Amt: sdkmath.NewInt(15)}, // 15 from val1 to val10
		{SrcValidator: "val1", DstValidator: "val9", Amt: sdkmath.NewInt(6)},   //  6 from val1 to val9

		{SrcValidator: "val2", DstValidator: "val9", Amt: sdkmath.NewInt(5)},  //  6 from val2 to val9
		{SrcValidator: "val2", DstValidator: "val8", Amt: sdkmath.NewInt(11)}, // 10 from val2 to val8
		{SrcValidator: "val2", DstValidator: "val7", Amt: sdkmath.NewInt(3)},  //  3 from val2 to val7

		{SrcValidator: "val3", DstValidator: "val7", Amt: sdkmath.NewInt(3)}, // 3 from val3 to val7
		{SrcValidator: "val3", DstValidator: "val6", Amt: sdkmath.NewInt(5)}, // 5 from val3 to val6
		{SrcValidator: "val3", DstValidator: "val5", Amt: sdkmath.NewInt(4)}, // 4 from val3 to val5
		{SrcValidator: "val3", DstValidator: "val4", Amt: sdkmath.NewInt(1)}, // 1 from val3 to val4
	}

	s.checkRebalanceICAMessages(validatorDeltas, expectedRebalancings)
}

func (s *KeeperTestSuite) TestGetRebalanceICAMessages_OddNumberValidators() {
	// Build up deltas for each validator, i.e. how much each validator needs to change by
	validatorDeltas := []keeper.RebalanceValidatorDelegationChange{
		// Overweight validators - they should lose some of their stake
		{ValidatorAddress: "val1", Delta: sdkmath.NewInt(15)}, // 15 to val11
		{ValidatorAddress: "val2", Delta: sdkmath.NewInt(12)}, // 6 to val11, 6 to val10
		{ValidatorAddress: "val3", Delta: sdkmath.NewInt(9)},  // 9 to val10
		{ValidatorAddress: "val4", Delta: sdkmath.NewInt(7)},  // 5 to val9, 2 to val8
		{ValidatorAddress: "val5", Delta: sdkmath.NewInt(2)},  // 2 to val8
		{ValidatorAddress: "val6", Delta: sdkmath.NewInt(2)},  // 2 to val7

		// Underweight validators - they should gain stake
		{ValidatorAddress: "val7", Delta: sdkmath.NewInt(-2)},   // 2 from val6
		{ValidatorAddress: "val8", Delta: sdkmath.NewInt(-4)},   // 2 from val4, 2 from val5
		{ValidatorAddress: "val9", Delta: sdkmath.NewInt(-5)},   // 5 from val4
		{ValidatorAddress: "val10", Delta: sdkmath.NewInt(-15)}, // 6 from val2, 9 from val3
		{ValidatorAddress: "val11", Delta: sdkmath.NewInt(-21)}, // 15 from val1, 6 from val2
	}

	// Build up the expected messages, moving across the list above
	expectedRebalancings := []types.Rebalancing{
		{SrcValidator: "val1", DstValidator: "val11", Amt: sdkmath.NewInt(15)}, // 15 from val1 to val11

		{SrcValidator: "val2", DstValidator: "val11", Amt: sdkmath.NewInt(6)}, // 6 from val2 to val11
		{SrcValidator: "val2", DstValidator: "val10", Amt: sdkmath.NewInt(6)}, // 6 from val2 to val10

		{SrcValidator: "val3", DstValidator: "val10", Amt: sdkmath.NewInt(9)}, // 9 from val3 to val10

		{SrcValidator: "val4", DstValidator: "val9", Amt: sdkmath.NewInt(5)}, // 5 from val4 to val9
		{SrcValidator: "val4", DstValidator: "val8", Amt: sdkmath.NewInt(2)}, // 2 from val4 to val8

		{SrcValidator: "val5", DstValidator: "val8", Amt: sdkmath.NewInt(2)}, // 2 from val5 to val8

		{SrcValidator: "val6", DstValidator: "val7", Amt: sdkmath.NewInt(2)}, // 2 from val6 to val7
	}

	s.checkRebalanceICAMessages(validatorDeltas, expectedRebalancings)
}

func (s *KeeperTestSuite) TestGetValidatorDelegationDifferences() {
	hostZone := types.HostZone{
		ChainId: HostChainId,
		Validators: []*types.Validator{
			// Total Weight: 100, Total Delegation: 200
			{Address: "val1", Weight: 10, Delegation: sdkmath.NewInt(20)},
			{Address: "val2", Weight: 20, Delegation: sdkmath.NewInt(140)},
			{Address: "val3", Weight: 70, Delegation: sdkmath.NewInt(40)},
			// Ignore this validator as it has a slash query in progresss
			{Address: "ignore", Weight: 50, Delegation: sdkmath.NewInt(100), SlashQueryInProgress: true},
		},
	}

	// Target delegation is determined by the total delegation * weight
	// Delta = Current - Target
	expectedDeltas := []keeper.RebalanceValidatorDelegationChange{
		// val1 is excluded because it's Target Delegation is equal to the Current Delegation (20)
		{ValidatorAddress: "val2", Delta: sdkmath.NewInt(140 - 40)}, // Current Delegation: 140, Target Delegation: 40
		{ValidatorAddress: "val3", Delta: sdkmath.NewInt(40 - 140)}, // Current Delegation: 40,  Target Delegation: 140
	}

	// Check delegation changes
	actualDeltas, err := s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, hostZone)
	s.Require().NoError(err, "no error expected when calculating delegation differences")
	s.Require().Len(actualDeltas, len(expectedDeltas), "number of redelegations")

	for i, expected := range expectedDeltas {
		s.Require().Equal(expected.ValidatorAddress, actualDeltas[i].ValidatorAddress, "address for delegation %d", i)
		s.Require().Equal(expected.Delta.Int64(), actualDeltas[i].Delta.Int64(), "delta for delegation %d", i)
	}

	// Check the error case when there are no delegations
	_, err = s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, types.HostZone{TotalDelegations: sdkmath.ZeroInt()})
	s.Require().ErrorContains(err, "unable to get target val amounts for host zone")
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone() {
	validators := []*types.Validator{
		{Address: "val1", Weight: 20},
		{Address: "val2", Weight: 40},
		{Address: "val3", Weight: 30},
		{Address: "val6", Weight: 5},
		{Address: "val5", Weight: 0},
		{Address: "val4", Weight: 5},
	}

	// Get targets with an even 100 total delegated - no overflow to last validator
	totalDelegation := sdkmath.NewInt(100)
	hostZone := types.HostZone{ChainId: HostChainId, Validators: validators}
	actualTargets, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, totalDelegation)
	s.Require().NoError(err, "no error expected when getting target weights for total delegation of 100")

	// Confirm target - should equal the validator's weight
	for _, validator := range validators {
		s.Require().Equal(int64(validator.Weight), actualTargets[validator.Address].Int64(),
			"validator %s target for total delegation of 100", validator.Address)
	}

	// Get targets with an uneven amount delegated - 77 - over flow to last validator
	totalDelegation = sdkmath.NewInt(77)
	expectedTargets := map[string]int64{
		"val5": 0,  // 0%  of 77 = 0
		"val4": 3,  // 5%  of 77 = 3.85 -> 3
		"val6": 3,  // 5%  of 77 = 3.85 -> 3
		"val1": 15, // 20% of 77 = 15.4 -> 15
		"val3": 23, // 30% of 77 = 23.1 -> 23
		"val2": 33, // Gets all overflow: 77 - 3 - 3 - 15 - 23 = 33
	}
	actualTargets, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, totalDelegation)
	s.Require().NoError(err, "no error expected when getting target weights for total delegation of 77")

	// Confirm target amounts again
	for validatorAddress, expectedTarget := range expectedTargets {
		s.Require().Equal(expectedTarget, actualTargets[validatorAddress].Int64(),
			"validator %s target for total delegation of 77", validatorAddress)
	}

	// Check zero delegations throws an error
	_, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, sdkmath.ZeroInt())
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is less than or equal to zero")

	// Check zero weights throws an error
	_, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, types.HostZone{}, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestGetTotalValidatorWeight() {
	validators := []types.Validator{
		{Address: "val1", Weight: 1},
		{Address: "val2", Weight: 2},
		{Address: "val3", Weight: 3},
		{Address: "val4", Weight: 4},
		{Address: "val5", Weight: 5},
	}
	expectedTotalWeights := int64(1 + 2 + 3 + 4 + 5)

	actualTotalWeight := s.App.StakeibcKeeper.GetTotalValidatorWeight(validators)

	s.Require().Equal(expectedTotalWeights, int64(actualTotalWeight))
}
