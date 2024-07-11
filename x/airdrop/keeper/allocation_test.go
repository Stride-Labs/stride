package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func (s *KeeperTestSuite) addUserAllocations() (userAllocations []types.UserAllocation) {
	for i := 0; i <= 4; i++ {
		userAllocation := types.UserAllocation{
			AirdropId: fmt.Sprintf("airdrop-%d", i),
			Address:   fmt.Sprintf("stride%d", i),
			Claimed:   sdkmath.ZeroInt(),
		}
		userAllocations = append(userAllocations, userAllocation)
		s.App.AirdropKeeper.SetUserAllocation(s.Ctx, userAllocation)
	}
	return userAllocations
}

func (s *KeeperTestSuite) TestGetUserAllocations() {
	userAllocations := s.addUserAllocations()

	for i := 0; i < len(userAllocations); i++ {
		expected := userAllocations[i]
		actual, found := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, expected.AirdropId, expected.Address)
		s.Require().True(found, "user allocation %s %s should have been found", expected.AirdropId, expected.Address)
		s.Require().Equal(expected, actual, "user allocation %s %s", expected.AirdropId, expected.Address)
	}
}

func (s *KeeperTestSuite) TestRemoveUserAllocation() {
	userAllocations := s.addUserAllocations()

	for removedIndex := 0; removedIndex < len(userAllocations); removedIndex++ {
		// Remove from removed index
		removedAllocation := userAllocations[removedIndex]
		s.App.AirdropKeeper.RemoveUserAllocation(s.Ctx, removedAllocation.AirdropId, removedAllocation.Address)

		// Confirm removed
		_, found := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, removedAllocation.AirdropId, removedAllocation.Address)
		s.Require().False(found, "user allocation %s %s should have been removed",
			removedAllocation.AirdropId, removedAllocation.Address)

		// Check all other user allocations are still there
		for checkedIndex := removedIndex + 1; checkedIndex < len(userAllocations); checkedIndex++ {
			checkedAllocation := userAllocations[checkedIndex]
			_, found := s.App.AirdropKeeper.GetUserAllocation(s.Ctx, checkedAllocation.AirdropId, checkedAllocation.Address)
			s.Require().True(found, "user allocation %s %d should still exist after removal of %s %s",
				checkedAllocation.AirdropId, checkedAllocation.Address, removedAllocation.AirdropId, removedAllocation.Address)
		}
	}
}

func (s *KeeperTestSuite) TestGetAllUserAllocations() {
	expectedUserAllocations := s.addUserAllocations()
	actualUserAllocations := s.App.AirdropKeeper.GetAllUserAllocations(s.Ctx)
	s.Require().Equal(len(expectedUserAllocations), len(actualUserAllocations), "number of user allocatinos")
	s.Require().ElementsMatch(expectedUserAllocations, actualUserAllocations)
}
