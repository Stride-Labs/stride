package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

type rebalanceExpectedValidatorDeltas struct {
	address         string
	balancedDelta   int64
	unbalancedDelta int64
}

func (s *KeeperTestSuite) TestGetRebalanceICAMessages() {

}

func (s *KeeperTestSuite) TestGetValidatorDelegationDifferences() {
	validators := []*types.Validator{
		// Total Weight: 100, Total Balance Delegation: 100, Total Unbalanced Delegation: 200
		{Address: "val1", Weight: 10, BalancedDelegation: sdkmath.NewInt(70), UnbalancedDelegation: sdkmath.NewInt(20)},
		{Address: "val2", Weight: 20, BalancedDelegation: sdkmath.NewInt(20), UnbalancedDelegation: sdkmath.NewInt(140)},
		{Address: "val3", Weight: 70, BalancedDelegation: sdkmath.NewInt(10), UnbalancedDelegation: sdkmath.NewInt(40)},
	}
	hostZone := types.HostZone{ChainId: HostChainId, Validators: validators}

	// Expected Balance is determined by the total delegation * weight
	// Delta = Expected - Current
	expectedValidatorDeltas := []rebalanceExpectedValidatorDeltas{
		{address: "val1", balancedDelta: 10 - 70, unbalancedDelta: 20 - 20},  // Expected Balanced: 10, Expected Unbalanced: 20
		{address: "val2", balancedDelta: 20 - 20, unbalancedDelta: 40 - 140}, // Expected Balanced: 20, Expected Unbalanced: 40
		{address: "val3", balancedDelta: 70 - 10, unbalancedDelta: 140 - 40}, // Expected Balanced: 70, Expected Unbalanced: 140
	}

	// Check delegation changes for the balanced portion
	actualBalancedDeltas, err := s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, hostZone, types.DelegationType_BALANCED)
	s.Require().NoError(err, "no error expected when calculated balanced delegation differences")
	for i, expected := range expectedValidatorDeltas {
		s.Require().Equal(expected.address, actualBalancedDeltas[i].ValidatorAddress, "address for balanced delegation %d", i)
		s.Require().Equal(expected.balancedDelta, actualBalancedDeltas[i].Delta.Int64(), "delta for balanced delegation %d", i)
	}

	// Check delegation changes for the unbalanced portion
	actualUnbalancedDeltas, err := s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, hostZone, types.DelegationType_UNBALANCED)
	s.Require().NoError(err, "no error expected when calculated unbalanced delegation differences")
	for i, expected := range expectedValidatorDeltas {
		s.Require().Equal(expected.address, actualUnbalancedDeltas[i].ValidatorAddress, "address for unbalanced delegation %d", i)
		s.Require().Equal(expected.unbalancedDelta, actualUnbalancedDeltas[i].Delta.Int64(), "delta for unbalanced delegation %d", i)
	}

	// Check the error case when there are no delegations
	_, err = s.App.StakeibcKeeper.GetValidatorDelegationDifferences(s.Ctx, types.HostZone{}, types.DelegationType_BALANCED)
	s.Require().ErrorContains(err, "unable to get target val amounts for host zone")
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone() {
	initialValidators := []*types.Validator{
		{Address: "val1", Weight: 20},
		{Address: "val2", Weight: 40},
		{Address: "val3", Weight: 30},
		{Address: "val4", Weight: 5},
		{Address: "val5", Weight: 0},
		{Address: "val6", Weight: 5},
	}
	expectedValidators := []*types.Validator{ // sorted by weight and name
		{Address: "val5", Weight: 0},
		{Address: "val4", Weight: 5},
		{Address: "val6", Weight: 5},
		{Address: "val1", Weight: 20},
		{Address: "val3", Weight: 30},
		{Address: "val2", Weight: 40},
	}

	// Get targets with an even 100 total delegated
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

	// Get targets with an uneven amount delegated - 77
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
		{Address: "val1", BalancedDelegation: sdkmath.NewInt(1), UnbalancedDelegation: sdkmath.NewInt(6)},
		{Address: "val2", BalancedDelegation: sdkmath.NewInt(2), UnbalancedDelegation: sdkmath.NewInt(7)},
		{Address: "val3", BalancedDelegation: sdkmath.NewInt(3), UnbalancedDelegation: sdkmath.NewInt(8)},
		{Address: "val4", BalancedDelegation: sdkmath.NewInt(4), UnbalancedDelegation: sdkmath.NewInt(9)},
		{Address: "val5", BalancedDelegation: sdkmath.NewInt(5), UnbalancedDelegation: sdkmath.NewInt(10)},
	}
	expectedBalancedDelegation := int64(1 + 2 + 3 + 4 + 5)
	expectedUnbalancedDelegation := int64(6 + 7 + 8 + 9 + 10)

	hostZone := types.HostZone{Validators: validators}
	actualBalancedDelegations := s.App.StakeibcKeeper.GetTotalValidatorDelegations(hostZone, types.DelegationType_BALANCED)
	actualUnbalancedDelegations := s.App.StakeibcKeeper.GetTotalValidatorDelegations(hostZone, types.DelegationType_UNBALANCED)

	s.Require().Equal(expectedBalancedDelegation, actualBalancedDelegations.Int64(), "balanced delegations")
	s.Require().Equal(expectedUnbalancedDelegation, actualUnbalancedDelegations.Int64(), "unbalanced delegations")
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
