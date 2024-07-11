package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func TestRemainingAllocations(t *testing.T) {
	// Filled allocations
	userAllocation := types.UserAllocation{
		Allocations: []sdkmath.Int{
			sdkmath.ZeroInt(),
			sdkmath.NewInt(1),
			sdkmath.NewInt(5),
			sdkmath.ZeroInt(),
			sdkmath.NewInt(3),
		},
	}
	require.Equal(t, int64(9), userAllocation.RemainingAllocations().Int64())

	// Populated array with all 0s
	userAllocation = types.UserAllocation{
		Allocations: []sdkmath.Int{
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
		},
	}
	require.Equal(t, int64(0), userAllocation.RemainingAllocations().Int64())

	// Uninitialized
	userAllocation = types.UserAllocation{}
	require.Equal(t, int64(0), userAllocation.RemainingAllocations().Int64())
}
