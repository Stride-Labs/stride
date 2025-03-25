package types

import sdkmath "cosmossdk.io/math"

// Calculates the remaining allocations for a user by summing each day
func (u UserAllocation) GetRemainingAllocations() sdkmath.Int {
	remaining := sdkmath.ZeroInt()
	for _, dailyAllocation := range u.Allocations {
		remaining = remaining.Add(dailyAllocation)
	}
	return remaining
}

// Calculates the eligible allocations for a user by summing up to
// the current date index
func (u UserAllocation) GetClaimableAllocation(currentDateIndex int) sdkmath.Int {
	if currentDateIndex > len(u.Allocations) {
		return sdkmath.ZeroInt()
	}

	claimable := sdkmath.ZeroInt()
	for i := 0; i <= currentDateIndex; i++ {
		claimable = claimable.Add(u.Allocations[i])
	}
	return claimable
}
