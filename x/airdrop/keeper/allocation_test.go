package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v25/x/airdrop/types"
)

// Creates a user allocation using the specified Ids and filling in default values
// for pointers
func newUserAllocation(airdropId, address string) types.UserAllocation {
	return types.UserAllocation{
		AirdropId: airdropId,
		Address:   address,
		Claimed:   sdkmath.ZeroInt(),
		Forfeited: sdkmath.ZeroInt(),
	}
}

func (s *KeeperTestSuite) addUserAllocations() (userAllocations []types.UserAllocation) {
	for i := 0; i <= 4; i++ {
		airdropId := fmt.Sprintf("airdrop-%d", i)
		address := fmt.Sprintf("stride%d", i)
		userAllocation := newUserAllocation(airdropId, address)

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

func (s *KeeperTestSuite) TestGetUserAllocationsForAirdrop() {
	// Define the expected allocations for each airdrop
	expectedAllocationsByAirdrop := map[string][]types.UserAllocation{
		"airdrop-1": {
			newUserAllocation("airdrop-1", "address-1"),
			newUserAllocation("airdrop-1", "address-2"),
			newUserAllocation("airdrop-1", "address-3"),
		},
		"airdrop-2": {
			newUserAllocation("airdrop-2", "address-1"),
			newUserAllocation("airdrop-2", "address-3"),
		},
		"airdrop-3": {
			newUserAllocation("airdrop-3", "address-1"),
			newUserAllocation("airdrop-3", "address-3"),
			newUserAllocation("airdrop-3", "address-4"),
		},
	}

	// Create each allocation
	for _, expectedAllocations := range expectedAllocationsByAirdrop {
		for _, expectedAllocation := range expectedAllocations {
			s.App.AirdropKeeper.SetUserAllocation(s.Ctx, expectedAllocation)
		}
	}

	// Grab the allocations for each airdrop
	for airdropId, expectedAllocations := range expectedAllocationsByAirdrop {
		actualAllocations := s.App.AirdropKeeper.GetUserAllocationsForAirdrop(s.Ctx, airdropId)
		s.Require().Equal(expectedAllocations, actualAllocations, "allocations for %s", airdropId)
	}
}

func (s *KeeperTestSuite) TestGetUserAllocationsForAddress() {
	// Define the expected allocations for each address
	expectedAllocationsByAddress := map[string][]types.UserAllocation{
		"address-1": {
			newUserAllocation("airdrop-1", "address-1"),
			newUserAllocation("airdrop-2", "address-1"),
			newUserAllocation("airdrop-3", "address-1"),
		},
		"address-2": {
			newUserAllocation("airdrop-1", "address-2"),
			newUserAllocation("airdrop-3", "address-2"),
		},
		"address-3": {
			newUserAllocation("airdrop-1", "address-3"),
			newUserAllocation("airdrop-3", "address-3"),
			newUserAllocation("airdrop-4", "address-3"),
		},
	}

	// Create each allocation and airdrop
	for _, allocations := range expectedAllocationsByAddress {
		for _, expectedAllocation := range allocations {
			s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{Id: expectedAllocation.AirdropId})
			s.App.AirdropKeeper.SetUserAllocation(s.Ctx, expectedAllocation)
		}
	}

	// Grab the allocations for each address
	for address, expectedAllocations := range expectedAllocationsByAddress {
		actualAllocations := s.App.AirdropKeeper.GetUserAllocationsForAddress(s.Ctx, address)
		s.Require().Equal(expectedAllocations, actualAllocations, "allocations for %s", address)
	}
}
