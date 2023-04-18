package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetRebalanceICAMessages() {

}

func (s *KeeperTestSuite) TestGetValidatorDelegationDifferences() {

}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone() {

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
