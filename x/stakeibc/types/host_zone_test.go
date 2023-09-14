package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func TestHostZoneUnbondingFrequency(t *testing.T) {
	testCases := []struct {
		unbondingPeriod    uint64
		unbondingFrequency uint64
	}{
		{
			unbondingPeriod:    1,
			unbondingFrequency: 1,
		},
		{
			unbondingPeriod:    6,
			unbondingFrequency: 1,
		},
		{
			unbondingPeriod:    7,
			unbondingFrequency: 2,
		},
		{
			unbondingPeriod:    13,
			unbondingFrequency: 2,
		},
		{
			unbondingPeriod:    14,
			unbondingFrequency: 3,
		},
		{
			unbondingPeriod:    20,
			unbondingFrequency: 3,
		},
		{
			unbondingPeriod:    21,
			unbondingFrequency: 4,
		},
		{
			unbondingPeriod:    27,
			unbondingFrequency: 4,
		},
		{
			unbondingPeriod:    28,
			unbondingFrequency: 5,
		},
	}

	for _, tc := range testCases {
		hostZone := types.HostZone{
			UnbondingPeriod: tc.unbondingPeriod,
		}
		require.Equal(t, tc.unbondingFrequency, hostZone.GetUnbondingFrequency(), "unbonding frequency")
	}
}
