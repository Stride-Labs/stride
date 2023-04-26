package types_test

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
)

func init() {
	apptesting.SetupConfig()
}

func getStakeibcMemo(address, action string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "stride_address": "%[1]s", "action": "%[2]s" } 
			}
		}`, address, action)
}

func getClaimMemo(address string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"claim": { "stride_address": "%[1]s" } 
			}
		}`, address)
}

func getClaimAndStakeibcMemo(address, action string) string {
	return fmt.Sprintf(`
	    {
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "stride_address": "%[1]s", "action": "%[2]s" },
				"claim": { "stride_address": "%[1]s" } 
			}
		}`, address, action)
}

// Helper function to check the routingInfo with a switch statement
// This isn't the most efficient way to check the type  (require.TypeOf could be used instead)
//  but it better aligns with how the routing info is checked in module_ibc
func checkModuleRoutingInfoType(routingInfo types.ModuleRoutingInfo, expectedType string) bool {
	switch routingInfo.(type) {
	case types.StakeibcPacketMetadata:
		return expectedType == "stakeibc"
	case types.ClaimPacketMetadata:
		return expectedType == "claim"
	default:
		return false
	}
}

func TestParsePacketMetadata(t *testing.T) {
	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validStakeibcAction := "LiquidStake"

	validParsedStakeibcPacketMetadata := types.StakeibcPacketMetadata{
		StrideAddress: validAddress,
		Action:        validStakeibcAction,
	}

	validParsedClaimPacketMetadata := types.ClaimPacketMetadata{
		StrideAddress: validAddress,
	}

	testCases := []struct {
		name                string
		metadata            string
		parsedStakeibc      *types.StakeibcPacketMetadata
		parsedClaim         *types.ClaimPacketMetadata
		expectedNilMetadata bool
		expectedErr         string
	}{
		{
			name:           "valid stakeibc memo",
			metadata:       getStakeibcMemo(validAddress, validStakeibcAction),
			parsedStakeibc: &validParsedStakeibcPacketMetadata,
		},
		{
			name:        "valid claim memo",
			metadata:    getClaimMemo(validAddress),
			parsedClaim: &validParsedClaimPacketMetadata,
		},
		{
			name:                "normal IBC transfer",
			metadata:            validAddress, // normal address - not autopilot JSON
			expectedNilMetadata: true,
		},
		{
			name:                "empty memo",
			metadata:            "",
			expectedNilMetadata: true,
		},
		{
			name:                "empty JSON memo",
			metadata:            "{}",
			expectedNilMetadata: true,
		},
		{
			name:                "different module specified",
			metadata:            `{ "other_module": { } }`,
			expectedNilMetadata: true,
		},
		{
			name:        "empty receiver address",
			metadata:    `{ "autopilot": { } }`,
			expectedErr: "receiver address must be specified when using autopilot",
		},
		{
			name:        "invalid receiver address",
			metadata:    `{ "autopilot": { "receiver": "invalid_address" } }`,
			expectedErr: "receiver address must be specified when using autopilot",
		},
		{
			name:        "invalid stakeibc address",
			metadata:    getStakeibcMemo(invalidAddress, validStakeibcAction),
			expectedErr: "receiver address must be specified when using autopilot",
		},
		{
			name:        "invalid stakeibc action",
			metadata:    getStakeibcMemo(validAddress, "bad_action"),
			expectedErr: "unsupported stakeibc action",
		},
		{
			name:        "invalid claim address",
			metadata:    getClaimMemo(invalidAddress),
			expectedErr: "receiver address must be specified when using autopilot",
		},
		{
			name:        "both claim and stakeibc memo set",
			metadata:    getClaimAndStakeibcMemo(validAddress, validStakeibcAction),
			expectedErr: "invalid number of module routes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedData, actualErr := types.ParsePacketMetadata(tc.metadata)

			if tc.expectedErr == "" {
				require.NoError(t, actualErr)
				if tc.expectedNilMetadata {
					require.Nil(t, parsedData, "parsed data response should be nil")
				} else {
					if tc.parsedStakeibc != nil {
						checkModuleRoutingInfoType(parsedData.RoutingInfo, "stakeibc")
						routingInfo, ok := parsedData.RoutingInfo.(types.StakeibcPacketMetadata)
						require.True(t, ok, "routing info should be stakeibc")
						require.Equal(t, *tc.parsedStakeibc, routingInfo, "parsed stakeibc value")
					} else if tc.parsedClaim != nil {
						checkModuleRoutingInfoType(parsedData.RoutingInfo, "claim")
						routingInfo, ok := parsedData.RoutingInfo.(types.ClaimPacketMetadata)
						require.True(t, ok, "routing info should be claim")
						require.Equal(t, *tc.parsedClaim, routingInfo, "parsed claim value")
					}
				}
			} else {
				require.ErrorContains(t, actualErr, types.ErrInvalidPacketMetadata.Error(), "expected error type for %s", tc.name)
				require.ErrorContains(t, actualErr, tc.expectedErr, "expected error for %s", tc.name)
			}
		})
	}
}

func TestValidateStakeibcPacketMetadata(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()
	validAction := "LiquidStake"

	testCases := []struct {
		name        string
		metadata    *types.StakeibcPacketMetadata
		expectedErr string
	}{
		{
			name: "valid Metadata data",
			metadata: &types.StakeibcPacketMetadata{
				StrideAddress: validAddress,
				Action:        validAction,
			},
		},
		{
			name: "invalid address",
			metadata: &types.StakeibcPacketMetadata{
				StrideAddress: "bad_address",
				Action:        validAction,
			},
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid action",
			metadata: &types.StakeibcPacketMetadata{
				StrideAddress: validAddress,
				Action:        "bad_action",
			},
			expectedErr: "unsupported stakeibc action",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := tc.metadata.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}

func TestValidateClaimPacketMetadata(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()

	testCases := []struct {
		name        string
		metadata    *types.ClaimPacketMetadata
		expectedErr string
	}{
		{
			name: "valid metadata",
			metadata: &types.ClaimPacketMetadata{
				StrideAddress: validAddress,
			},
		},
		{
			name: "invalid address",
			metadata: &types.ClaimPacketMetadata{
				StrideAddress: "bad_address",
			},
			expectedErr: "decoding bech32 failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := tc.metadata.Validate()
			if tc.expectedErr == "" {
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}
