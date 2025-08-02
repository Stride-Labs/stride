package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
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

func TestSafelyGetCommunityPoolRebate(t *testing.T) {
	chainId := "chain-0"

	testCases := []struct {
		name           string
		hostZone       types.HostZone
		expectedRebate bool
	}{
		{
			name:           "no rebate",
			hostZone:       types.HostZone{ChainId: chainId},
			expectedRebate: false,
		},
		{
			name: "rebate but empty percentage field",
			hostZone: types.HostZone{
				ChainId: chainId,
				CommunityPoolRebate: &types.CommunityPoolRebate{
					LiquidStakedStTokenAmount: sdkmath.NewInt(1),
				},
			},
			expectedRebate: false,
		},
		{
			name: "rebate but empty liquid stake amount",
			hostZone: types.HostZone{
				ChainId: chainId,
				CommunityPoolRebate: &types.CommunityPoolRebate{
					RebateRate: sdkmath.LegacyOneDec(),
				},
			},
			expectedRebate: false,
		},
		{
			name: "valid rebate",
			hostZone: types.HostZone{
				ChainId: chainId,
				CommunityPoolRebate: &types.CommunityPoolRebate{
					RebateRate:                sdkmath.LegacyOneDec(),
					LiquidStakedStTokenAmount: sdkmath.NewInt(1),
				},
			},
			expectedRebate: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualRebate, hasRebate := tc.hostZone.SafelyGetCommunityPoolRebate()
			require.Equal(t, tc.expectedRebate, hasRebate, "has rebate bool")

			if tc.expectedRebate {
				require.Equal(t, *tc.hostZone.CommunityPoolRebate, actualRebate, "rebate")
			}
		})
	}
}
