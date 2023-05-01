package keeper_test

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Given a set of validator deltas (containing the expected change in delegation for each validator)
// and a set of expected rebalancings (containing the individual rebalance messages), calls
// RebalanceICAMessages and checks that the corresponding ICA messages match the expected rebalancings
func (s *KeeperTestSuite) checkRebalanceICAMessages(
	validatorDeltas []keeper.RebalanceValidatorDelegationChange,
	expectedRebalancings []types.Rebalancing,
) {
	// Build the expected ICA messages from the list of rebalancings above
	delegationAddress := "cosmos_DELEGATION"
	expectedMsgs := []sdk.Msg{}
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
	validators := []*types.Validator{
		// Total Weight: 100, Total Delegation: 200
		{Address: "val1", Weight: 10, Delegation: sdkmath.NewInt(20)},
		{Address: "val2", Weight: 20, Delegation: sdkmath.NewInt(140)},
		{Address: "val3", Weight: 70, Delegation: sdkmath.NewInt(40)},
	}

	// Shuffle the validators to ensure the sorting worked
	rand.Shuffle(len(validators), func(i, j int) {
		validators[i], validators[j] = validators[j], validators[i]
	})
	hostZone := types.HostZone{ChainId: HostChainId, Validators: validators}

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
	_, err = s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, types.HostZone{})
	s.Require().ErrorContains(err, "unable to get target val amounts for host zone")
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone() {
	initialValidators := []*types.Validator{
		{Address: "val1", Weight: 20},
		{Address: "val2", Weight: 40},
		{Address: "val3", Weight: 30},
		{Address: "val6", Weight: 5},
		{Address: "val5", Weight: 0},
		{Address: "val4", Weight: 5},
	}
	expectedValidators := []*types.Validator{ // sorted by weight and name
		{Address: "val5", Weight: 0},
		{Address: "val4", Weight: 5},
		{Address: "val6", Weight: 5},
		{Address: "val1", Weight: 20},
		{Address: "val3", Weight: 30},
		{Address: "val2", Weight: 40},
	}

	// Get targets with an even 100 total delegated - no overflow to last validator
	totalDelegation := sdkmath.NewInt(100)
	hostZone := types.HostZone{ChainId: HostChainId, Validators: initialValidators}
	actualTargets, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, totalDelegation)
	s.Require().NoError(err, "no error expected when getting target weights for total delegation of 100")

	// Confirm new validator ordering (we check the original host zone because the re-ordering is in place)
	for i := 0; i < len(hostZone.Validators); i++ {
		s.Require().Equal(expectedValidators[i].Address, hostZone.Validators[i].Address,
			"validator %d weight", i)
	}

	// Confirm target - should equal the validator's weight
	for _, expectedValidators := range expectedValidators {
		s.Require().Equal(int64(expectedValidators.Weight), actualTargets[expectedValidators.Address].Int64(),
			"validator %s target for total delegation of 100", expectedValidators.Address)
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
	s.Require().ErrorContains(err, "Cannot calculate target delegation if final amount is 0")

	// Check zero weights throws an error
	_, err = s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, types.HostZone{}, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "No non-zero validators found for host zone")
}

func (s *KeeperTestSuite) TestGetTotalValidatorDelegations() {
	validators := []*types.Validator{
		{Address: "val1", Delegation: sdkmath.NewInt(1)},
		{Address: "val2", Delegation: sdkmath.NewInt(2)},
		{Address: "val3", Delegation: sdkmath.NewInt(3)},
		{Address: "val4", Delegation: sdkmath.NewInt(4)},
		{Address: "val5", Delegation: sdkmath.NewInt(5)},
	}
	expectedDelegation := int64(1 + 2 + 3 + 4 + 5)

	hostZone := types.HostZone{Validators: validators}
	actualDelegations := s.App.StakeibcKeeper.GetTotalValidatorDelegations(hostZone)

	s.Require().Equal(expectedDelegation, actualDelegations.Int64(), "delegations")
}

func (s *KeeperTestSuite) TestGetTotalValidatorWeight() {
	validators := []*types.Validator{
		{Address: "val1", Weight: 1},
		{Address: "val2", Weight: 2},
		{Address: "val3", Weight: 3},
		{Address: "val4", Weight: 4},
		{Address: "val5", Weight: 5},
	}
	expectedTotalWeights := int64(1 + 2 + 3 + 4 + 5)

	hostZone := types.HostZone{Validators: validators}
	actualTotalWeight := s.App.StakeibcKeeper.GetTotalValidatorWeight(hostZone)

	s.Require().Equal(expectedTotalWeights, int64(actualTotalWeight))
}
