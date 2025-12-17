package types_test

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v31/app/apptesting"
	"github.com/Stride-Labs/stride/v31/x/staketia/types"
)

const (
	Uninitialized    = "uninitialized"
	UninitializedInt = 999
)

// Helper function to fill in individual fields
// If the field is empty, replace with a valid value
// If the field is "uninitialized", replace with an empty value
func fillDefaultValue(currentValue, defaultValidValue string) string {
	if currentValue == "" {
		return defaultValidValue
	} else if currentValue == Uninitialized {
		return ""
	}
	return currentValue
}

// Helper function to fill in default values for the host zone struct
// if they're not specified in the test case
func fillDefaultHostZone(hostZone types.HostZone) types.HostZone {
	validChannelId := "channel-0"
	validDenom := "denom"

	hostZone.ChainId = fillDefaultValue(hostZone.ChainId, "chain-0")
	hostZone.TransferChannelId = fillDefaultValue(hostZone.TransferChannelId, validChannelId)
	hostZone.NativeTokenDenom = fillDefaultValue(hostZone.NativeTokenDenom, validDenom)

	ibcDenomTracePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, validChannelId)
	defaultIbcDenom := transfertypes.ParseDenomTrace(ibcDenomTracePrefix + validDenom).IBCDenom()
	hostZone.NativeTokenIbcDenom = fillDefaultValue(hostZone.NativeTokenIbcDenom, defaultIbcDenom)

	hostZone.DelegationAddress = fillDefaultValue(hostZone.DelegationAddress, "celestiaXXX")
	hostZone.RewardAddress = fillDefaultValue(hostZone.RewardAddress, "celestiaXXX")

	validAddress := apptesting.CreateRandomAccounts(1)[0].String()
	hostZone.DepositAddress = fillDefaultValue(hostZone.DepositAddress, validAddress)
	hostZone.RedemptionAddress = fillDefaultValue(hostZone.RedemptionAddress, validAddress)
	hostZone.ClaimAddress = fillDefaultValue(hostZone.ClaimAddress, validAddress)
	hostZone.OperatorAddressOnStride = fillDefaultValue(hostZone.OperatorAddressOnStride, validAddress)
	hostZone.SafeAddressOnStride = fillDefaultValue(hostZone.SafeAddressOnStride, validAddress)

	if hostZone.UnbondingPeriodSeconds == UninitializedInt {
		hostZone.UnbondingPeriodSeconds = 0 // invalid
	} else {
		hostZone.UnbondingPeriodSeconds = 21 // valid
	}

	return hostZone
}

func TestValidateHostZoneGenesis(t *testing.T) {
	// For each test case, assume all excluded values are valid
	// (they'll be filled in downstream)
	testCases := []struct {
		name          string
		hostZone      types.HostZone
		expectedError string
	}{
		{
			name:     "valid host zone",
			hostZone: types.HostZone{},
		},
		{
			name: "missing chain-id",
			hostZone: types.HostZone{
				ChainId: Uninitialized,
			},
			expectedError: "chain-id must be specified",
		},
		{
			name: "missing transfer channel-id",
			hostZone: types.HostZone{
				TransferChannelId: Uninitialized,
			},
			expectedError: "transfer channel-id must be specified",
		},
		{
			name: "missing token denom",
			hostZone: types.HostZone{
				NativeTokenDenom: Uninitialized,
			},
			expectedError: "native token denom must be specified",
		},
		{
			name: "missing token ibc denom",
			hostZone: types.HostZone{
				NativeTokenIbcDenom: Uninitialized,
			},
			expectedError: "native token ibc denom must be specified",
		},
		{
			name: "ibc denom mismatch",
			hostZone: types.HostZone{
				NativeTokenIbcDenom: "ibc/XXX",
			},
			expectedError: "native token ibc denom did not match hash generated",
		},
		{
			name: "missing delegation address",
			hostZone: types.HostZone{
				DelegationAddress: Uninitialized,
			},
			expectedError: "delegation address must be specified",
		},
		{
			name: "missing reward address",
			hostZone: types.HostZone{
				RewardAddress: Uninitialized,
			},
			expectedError: "reward address must be specified",
		},
		{
			name: "missing deposit address",
			hostZone: types.HostZone{
				DepositAddress: Uninitialized,
			},
			expectedError: "deposit address must be specified",
		},
		{
			name: "missing redemption address",
			hostZone: types.HostZone{
				RedemptionAddress: Uninitialized,
			},
			expectedError: "redemption address must be specified",
		},
		{
			name: "missing claim address",
			hostZone: types.HostZone{
				ClaimAddress: Uninitialized,
			},
			expectedError: "claim address must be specified",
		},
		{
			name: "missing operator address",
			hostZone: types.HostZone{
				OperatorAddressOnStride: Uninitialized,
			},
			expectedError: "operator address must be specified",
		},
		{
			name: "missing safe address",
			hostZone: types.HostZone{
				SafeAddressOnStride: Uninitialized,
			},
			expectedError: "safe address must be specified",
		},
		{
			name: "invalid deposit address",
			hostZone: types.HostZone{
				DepositAddress: "invalid_address",
			},
			expectedError: "invalid deposit address",
		},
		{
			name: "invalid redemption address",
			hostZone: types.HostZone{
				RedemptionAddress: "invalid_address",
			},
			expectedError: "invalid redemption address",
		},
		{
			name: "invalid claim address",
			hostZone: types.HostZone{
				ClaimAddress: "invalid_address",
			},
			expectedError: "invalid claim address",
		},
		{
			name: "invalid operator address",
			hostZone: types.HostZone{
				OperatorAddressOnStride: "invalid_address",
			},
			expectedError: "invalid operator address",
		},
		{
			name: "invalid safe address",
			hostZone: types.HostZone{
				SafeAddressOnStride: "invalid_address",
			},
			expectedError: "invalid safe address",
		},
		{
			name: "missing unbonding period",
			hostZone: types.HostZone{
				UnbondingPeriodSeconds: UninitializedInt,
			},
			expectedError: "unbonding period must be set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hostZone := fillDefaultHostZone(tc.hostZone)
			err := hostZone.ValidateGenesis()

			if tc.expectedError == "" {
				require.NoError(t, err, "no error expected")
			} else {
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
