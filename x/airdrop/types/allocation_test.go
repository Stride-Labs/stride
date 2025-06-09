package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v27/x/airdrop/types"
)

func TestGetRemainingAllocations(t *testing.T) {
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
	require.Equal(t, int64(9), userAllocation.GetRemainingAllocations().Int64())

	// Populated array with all 0s
	userAllocation = types.UserAllocation{
		Allocations: []sdkmath.Int{
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
		},
	}
	require.Equal(t, int64(0), userAllocation.GetRemainingAllocations().Int64())

	// Uninitialized
	userAllocation = types.UserAllocation{}
	require.Equal(t, int64(0), userAllocation.GetRemainingAllocations().Int64())
}

func TestGetClaimableAllocations(t *testing.T) {
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
	require.Equal(t, int64(0), userAllocation.GetClaimableAllocation(0).Int64())
	require.Equal(t, int64(1), userAllocation.GetClaimableAllocation(1).Int64())
	require.Equal(t, int64(6), userAllocation.GetClaimableAllocation(2).Int64())
	require.Equal(t, int64(6), userAllocation.GetClaimableAllocation(3).Int64())
	require.Equal(t, int64(9), userAllocation.GetClaimableAllocation(4).Int64())
	require.Equal(t, int64(0), userAllocation.GetClaimableAllocation(10).Int64()) // index out of bounds

	// Populated array with all 0s
	userAllocation = types.UserAllocation{
		Allocations: []sdkmath.Int{
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
			sdkmath.ZeroInt(),
		},
	}
	require.Equal(t, int64(0), userAllocation.GetClaimableAllocation(2).Int64())

	// Uninitialized
	userAllocation = types.UserAllocation{}
	require.Equal(t, int64(0), userAllocation.GetClaimableAllocation(3).Int64())
}
