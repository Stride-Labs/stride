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
	return fmt.Sprintf(`{"stakeibc": { "stride_address": "%s", "action": "%s" } }`, address, action)
}

func getClaimMemo(address, airdropId string) string {
	return fmt.Sprintf(`{"claim": { "stride_address": "%s", "airdrop_id": "%s" } }`, address, airdropId)
}

func getClaimAndStakeibcMemo(address, action, airdropId string) string {
	return fmt.Sprintf(`
	    {
			"stakeibc": { "stride_address": "%[1]s", "action": "%[2]s" },
			"claim": { "stride_address": "%[1]s", "airdrop_id": "%[3]s" } 
		}`, address, action, airdropId)
}

func TestParseReceiver(t *testing.T) {
	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validStakeibcAction := "LiquidStake"
	validAirdropId := "gaia"

	validParsedStakeibcReceiver := types.ParsedStakeibcReceiver{
		Enabled:       true,
		StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
		Action:        validStakeibcAction,
	}
	disabledStakeibcReceiver := types.ParsedStakeibcReceiver{
		Enabled: false,
	}

	validParsedClaimReceiver := types.ParsedClaimReceiver{
		Enabled:       true,
		StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
		AirdropId:     validAirdropId,
	}
	disabledClaimReceiver := types.ParsedClaimReceiver{
		Enabled: false,
	}

	testCases := []struct {
		name           string
		receivedData   string
		parsedStakeibc types.ParsedStakeibcReceiver
		parsedClaim    types.ParsedClaimReceiver
		expectedErr    string
	}{
		{
			name:           "valid stakeibc memo",
			receivedData:   getStakeibcMemo(validAddress, validStakeibcAction),
			parsedStakeibc: validParsedStakeibcReceiver,
			parsedClaim:    disabledClaimReceiver,
		},
		{
			name:           "valid claim memo",
			receivedData:   getClaimMemo(validAddress, validAirdropId),
			parsedStakeibc: disabledStakeibcReceiver,
			parsedClaim:    validParsedClaimReceiver,
		},
		{
			name:           "valid claim and stakeibc memo",
			receivedData:   getClaimAndStakeibcMemo(validAddress, validStakeibcAction, validAirdropId),
			parsedStakeibc: validParsedStakeibcReceiver,
			parsedClaim:    validParsedClaimReceiver,
		},
		{
			name:           "no messages",
			receivedData:   "{}",
			parsedStakeibc: disabledStakeibcReceiver,
			parsedClaim:    disabledClaimReceiver,
		},
		{
			name:           "no message - empty",
			receivedData:   "",
			parsedStakeibc: disabledStakeibcReceiver,
			parsedClaim:    disabledClaimReceiver,
		},
		{
			name:           "invalid memo",
			receivedData:   "bad_memo",
			parsedStakeibc: disabledStakeibcReceiver,
			parsedClaim:    disabledClaimReceiver,
			expectedErr:    "invalid character",
		},
		{
			name:         "invalid stakeibc address",
			receivedData: getStakeibcMemo(invalidAddress, validStakeibcAction),
			expectedErr:  "unknown address",
		},
		{
			name:         "invalid stakeibc action",
			receivedData: getStakeibcMemo(validAddress, "bad_action"),
			expectedErr:  "unsupported stakeibc action",
		},
		{
			name:         "invalid claim address",
			receivedData: getClaimMemo(invalidAddress, validAirdropId),
			expectedErr:  "unknown address",
		},
		{
			name:         "invalid claim airdrop",
			receivedData: getClaimMemo(validAddress, ""),
			expectedErr:  "invalid claim airdrop ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsedData, actualErr := types.ParseReceiverData(tc.receivedData)
			if tc.expectedErr == "" {
				require.NoError(t, actualErr)
				require.Equal(t, tc.parsedStakeibc, parsedData.Stakeibc, "parsed stakeibc")
				require.Equal(t, tc.parsedClaim, parsedData.Claim, "parsed claim")
			} else {
				require.ErrorContains(t, actualErr, types.ErrInvalidReceiverData.Error(), "expected error type for %s", tc.name)
				require.ErrorContains(t, actualErr, tc.expectedErr, "expected error for %s", tc.name)
			}
		})
	}
}

func TestParseStakeibcReceiverData(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()
	validAction := "LiquidStake"

	testCases := []struct {
		name           string
		raw            *types.RawStakeibcReceiver
		expectedParsed types.ParsedStakeibcReceiver
		expectedErr    string
	}{
		{
			name: "valid receiver data",
			raw: &types.RawStakeibcReceiver{
				StrideAddress: validAddress,
				Action:        validAction,
			},
			expectedParsed: types.ParsedStakeibcReceiver{
				StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
				Action:        validAction,
				Enabled:       true,
			},
		},
		{
			name: "empty raw message",
			raw:  nil,
			expectedParsed: types.ParsedStakeibcReceiver{
				Enabled: false,
			},
		},
		{
			name: "invalid address",
			raw: &types.RawStakeibcReceiver{
				StrideAddress: "bad_address",
				Action:        validAction,
			},
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid action",
			raw: &types.RawStakeibcReceiver{
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
				require.Equal(t, tc.expectedParsed, *actualParsed, "parsed receiver")
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}

func TestParseClaimReceiverData(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()
	validAirdropId := "gaia"

	testCases := []struct {
		name           string
		raw            *types.RawClaimReceiver
		expectedParsed types.ParsedClaimReceiver
		expectedErr    string
	}{
		{
			name: "valid receiver data",
			raw: &types.RawClaimReceiver{
				StrideAddress: validAddress,
				AirdropId:     validAirdropId,
			},
			expectedParsed: types.ParsedClaimReceiver{
				StrideAddress: sdk.MustAccAddressFromBech32(validAddress),
				AirdropId:     validAirdropId,
				Enabled:       true,
			},
		},
		{
			name: "empty raw message",
			raw:  nil,
			expectedParsed: types.ParsedClaimReceiver{
				Enabled: false,
			},
		},
		{
			name: "invalid address",
			raw: &types.RawClaimReceiver{
				StrideAddress: "bad_address",
				AirdropId:     validAirdropId,
			},
			expectedErr: "decoding bech32 failed",
		},
		{
			name: "invalid airdrop-id",
			raw: &types.RawClaimReceiver{
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
				require.Equal(t, tc.expectedParsed, *actualParsed, "parsed receiver")
				require.NoError(t, actualErr, "no error expected for %s", tc.name)
			} else {
				require.ErrorContains(t, actualErr, tc.expectedErr, "error expected for %s", tc.name)
			}
		})
	}
}
