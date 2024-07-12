package types

import sdkmath "cosmossdk.io/math"

// Calculates the remaining allocations for a user by summing each day
func (u UserAllocation) RemainingAllocations() sdkmath.Int {
	remaining := sdkmath.ZeroInt()
	for _, dailyAllocation := range u.Allocations {
		remaining = remaining.Add(dailyAllocation)
	}
	return remaining
}
