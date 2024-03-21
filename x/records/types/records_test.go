package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v20/x/records/types"
)

func TestShouldInitiateUnbonding(t *testing.T) {
	testCases := []struct {
		name         string
		status       types.HostZoneUnbonding_Status
		amount       sdkmath.Int
		shouldUnbond bool
	}{
		{
			name:         "should unbond",
			status:       types.HostZoneUnbonding_UNBONDING_QUEUE,
			amount:       sdkmath.NewInt(10),
			shouldUnbond: true,
		},
		{
			name:         "not in unbonding queue",
			status:       types.HostZoneUnbonding_CLAIMABLE,
			amount:       sdkmath.NewInt(10),
			shouldUnbond: false,
		},
		{
			name:         "zero amount",
			status:       types.HostZoneUnbonding_CLAIMABLE,
			amount:       sdkmath.ZeroInt(),
			shouldUnbond: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record := types.HostZoneUnbonding{
				Status:            tc.status,
				NativeTokenAmount: tc.amount,
			}
			require.Equal(t, tc.shouldUnbond, record.ShouldInitiateUnbonding())
		})
	}
}
