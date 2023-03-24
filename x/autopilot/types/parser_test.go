package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v7/app/apptesting"
	"github.com/Stride-Labs/stride/v7/x/autopilot/types"
)

func init() {
	apptesting.SetupConfig()
}

func getStakeibcMemo(address, action string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"stakeibc": { "stride_address": "%s", "action": "%s" } 
			}
		}`, address, action)
}

func getClaimMemo(address, airdropId string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"claim": { "stride_address": "%s", "airdrop_id": "%s" } 
			}
		}`, address, airdropId)
}

func getClaimAndStakeibcMemo(address, action, airdropId string) string {
	return fmt.Sprintf(`
	    {
			"autopilot": {
				"stakeibc": { "stride_address": "%[1]s", "action": "%[2]s" },
				"claim": { "stride_address": "%[1]s", "airdrop_id": "%[3]s" } 
			}
		}`, address, action, airdropId)
}

func TestParsePacketMetadata(t *testing.T) {
	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validStakeibcAction := "LiquidStake"
	validAirdropId := "gaia"

	validParsedStakeibcPacketMetadata := types.StakeibcPacketMetadata{
		Enabled:       true,
		StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
		Action:        validStakeibcAction,
	}
	disabledStakeibcPacketMetadata := types.StakeibcPacketMetadata{
		Enabled: false,
	}

	validParsedClaimPacketMetadata := types.ClaimPacketMetadata{
		Enabled:       true,
		StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
		AirdropId:     validAirdropId,
	}
	disabledClaimPacketMetadata := types.ClaimPacketMetadata{
		Enabled: false,
	}

	testCases := []struct {
		name                string
		metadata            string
		parsedMetadata      types.PacketMetadata
		parsedStakeibc      types.StakeibcPacketMetadata
		parsedClaim         types.ClaimPacketMetadata
		expectedNilMetadata bool
		expectedErr         string
	}{
		{
			name:           "valid stakeibc memo",
			metadata:       getStakeibcMemo(validAddress, validStakeibcAction),
			parsedStakeibc: validParsedStakeibcPacketMetadata,
			parsedClaim:    disabledClaimPacketMetadata,
		},
		{
			name:           "valid claim memo",
			metadata:       getClaimMemo(validAddress, validAirdropId),
			parsedStakeibc: disabledStakeibcPacketMetadata,
			parsedClaim:    validParsedClaimPacketMetadata,
		},
		{
			name:                "normal IBC transfer",
			metadata:            validAddress, // normal address - not autopilot JSON
			expectedNilMetadata: true,
		},
		{
			name:                "no messages",
			metadata:            "{}",
			parsedStakeibc:      disabledStakeibcPacketMetadata,
			parsedClaim:         disabledClaimPacketMetadata,
			expectedNilMetadata: true,
		},
		{
			name:                "no message - empty",
			metadata:            "",
			parsedStakeibc:      disabledStakeibcPacketMetadata,
			parsedClaim:         disabledClaimPacketMetadata,
			expectedNilMetadata: true,
		},
		{
			name:        "invalid stakeibc address",
			metadata:    getStakeibcMemo(invalidAddress, validStakeibcAction),
			expectedErr: "unknown address",
		},
		{
			name:        "invalid stakeibc action",
			metadata:    getStakeibcMemo(validAddress, "bad_action"),
			expectedErr: "unsupported stakeibc action",
		},
		{
			name:        "invalid claim address",
			metadata:    getClaimMemo(invalidAddress, validAirdropId),
			expectedErr: "unknown address",
		},
		{
			name:        "invalid claim airdrop",
			metadata:    getClaimMemo(validAddress, ""),
			expectedErr: "invalid claim airdrop ID",
		},
		{
			name:        "both claim and stakeibc memo set",
			metadata:    getClaimAndStakeibcMemo(validAddress, validStakeibcAction, validAirdropId),
			expectedErr: "multiple autopilot routes in the same transaction",
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
					require.Equal(t, tc.parsedStakeibc, parsedData.Stakeibc, "parsed stakeibc")
					require.Equal(t, tc.parsedClaim, parsedData.Claim, "parsed claim")
				}
			} else {
				require.ErrorContains(t, actualErr, types.ErrInvalidPacketMetadata.Error(), "expected error type for %s", tc.name)
				require.ErrorContains(t, actualErr, tc.expectedErr, "expected error for %s", tc.name)
			}
		})
	}
}

func TestParseStakeibcMetadataData(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()
	validAction := "LiquidStake"

	testCases := []struct {
		name           string
		raw            *types.RawStakeibcPacketMetadata
		expectedParsed types.StakeibcPacketMetadata
		expectedErr    string
	}{
		{
			name: "valid Metadata data",
			raw: &types.RawStakeibcPacketMetadata{
				StrideAddress: validAddress,
				Action:        validAction,
			},
			expectedParsed: types.StakeibcPacketMetadata{
				StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
				Action:        validAction,
				Enabled:       true,
			},
		},
		{
			name: "empty raw message",
			raw:  nil,
			expectedParsed: types.StakeibcPacketMetadata{
				Enabled: false,
			},
		},
		{
			name: "invalid address",
			raw: &types.RawStakeibcPacketMetadata{
				StrideAddress: "bad_address",
				Action:        validAction,
			},
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid action",
			raw: &types.RawStakeibcPacketMetadata{
				StrideAddress: validAddress,
				Action:        "bad_action",
			},
			expectedErr: "unsupported stakeibc action",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualParsed, actualErr := tc.raw.ParseAndValidate()
			if tc.expectedErr == "" {
				require.Equal(t, tc.expectedParsed, *actualParsed, "parsed metadata")
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}

func TestParseClaimMetadataData(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()
	validAirdropId := "gaia"

	testCases := []struct {
		name           string
		raw            *types.RawClaimPacketMetadata
		expectedParsed types.ClaimPacketMetadata
		expectedErr    string
	}{
		{
			name: "valid metadata",
			raw: &types.RawClaimPacketMetadata{
				StrideAddress: validAddress,
				AirdropId:     validAirdropId,
			},
			expectedParsed: types.ClaimPacketMetadata{
				StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
				AirdropId:     validAirdropId,
				Enabled:       true,
			},
		},
		{
			name: "empty raw message",
			raw:  nil,
			expectedParsed: types.ClaimPacketMetadata{
				Enabled: false,
			},
		},
		{
			name: "invalid address",
			raw: &types.RawClaimPacketMetadata{
				StrideAddress: "bad_address",
				AirdropId:     validAirdropId,
			},
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid airdrop-id",
			raw: &types.RawClaimPacketMetadata{
				StrideAddress: validAddress,
				AirdropId:     "",
			},
			expectedErr: "invalid claim airdrop ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualParsed, actualErr := tc.raw.ParseAndValidate()
			if tc.expectedErr == "" {
				require.Equal(t, tc.expectedParsed, *actualParsed, "parsed metadata")
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}
