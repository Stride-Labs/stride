package keeper_test

import (
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

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
