package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// ----------------------------------------------
//               MsgClaimDaily
// ----------------------------------------------

func TestMsgClaimDaily_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAirdropId := "airdrop-1"

	tests := []struct {
		name          string
		msg           types.MsgClaimDaily
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgClaimDaily{
				Claimer:   validAddress,
				AirdropId: validAirdropId,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgClaimDaily{
				Claimer:   invalidAddress,
				AirdropId: validAirdropId,
			},
			expectedError: "invalid address",
		},
		{
			name: "invalid address",
			msg: types.MsgClaimDaily{
				Claimer:   validAddress,
				AirdropId: "",
			},
			expectedError: "airdrop-id must be specified",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.msg.ValidateBasic()
			if tc.expectedError != "" {
				require.ErrorContains(t, actualError, tc.expectedError)
				return
			}
			require.NoError(t, actualError)
		})
	}
}

func TestMsgClaimDaily_GetSignBytes(t *testing.T) {
	addr := "strideXXX"
	airdropId := "airdrop"
	msg := types.NewMsgClaimDaily(addr, airdropId)
	res := msg.GetSignBytes()

	expected := `{"type":"airdrop/MsgClaimDaily","value":{"airdrop_id":"airdrop","claimer":"strideXXX"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgClaimEarly
// ----------------------------------------------

func TestMsgClaimEarly_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAirdropId := "airdrop-1"

	tests := []struct {
		name          string
		msg           types.MsgClaimEarly
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgClaimEarly{
				Claimer:   validAddress,
				AirdropId: validAirdropId,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgClaimEarly{
				Claimer:   invalidAddress,
				AirdropId: validAirdropId,
			},
			expectedError: "invalid address",
		},
		{
			name: "invalid address",
			msg: types.MsgClaimEarly{
				Claimer:   validAddress,
				AirdropId: "",
			},
			expectedError: "airdrop-id must be specified",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.msg.ValidateBasic()
			if tc.expectedError != "" {
				require.ErrorContains(t, actualError, tc.expectedError)
				return
			}
			require.NoError(t, actualError)
		})
	}
}

func TestMsgClaimEarly_GetSignBytes(t *testing.T) {
	addr := "strideXXX"
	airdropId := "airdrop"
	msg := types.NewMsgClaimEarly(addr, airdropId)
	res := msg.GetSignBytes()

	expected := `{"type":"airdrop/MsgClaimEarly","value":{"airdrop_id":"airdrop","claimer":"strideXXX"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgClaimAndStake
// ----------------------------------------------

func TestMsgClaimAndStake_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()

	validAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAirdropId := "airdrop-1"
	validValidatorAddress := "stridevaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rcfud5c"

	tests := []struct {
		name          string
		msg           types.MsgClaimAndStake
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgClaimAndStake{
				Claimer:          validAddress,
				AirdropId:        validAirdropId,
				ValidatorAddress: validValidatorAddress,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgClaimAndStake{
				Claimer:          invalidAddress,
				AirdropId:        validAirdropId,
				ValidatorAddress: validValidatorAddress,
			},
			expectedError: "invalid address",
		},
		{
			name: "invalid address",
			msg: types.MsgClaimAndStake{
				Claimer:          validAddress,
				AirdropId:        "",
				ValidatorAddress: validValidatorAddress,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "invalid address",
			msg: types.MsgClaimAndStake{
				Claimer:          validAddress,
				AirdropId:        validAirdropId,
				ValidatorAddress: "stridevaloper17kht2x2p",
			},
			expectedError: "invalid validator address",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.msg.ValidateBasic()
			if tc.expectedError != "" {
				require.ErrorContains(t, actualError, tc.expectedError)
				return
			}
			require.NoError(t, actualError)
		})
	}
}

func TestMsgClaimAndStake_GetSignBytes(t *testing.T) {
	addr := "strideXXX"
	airdropId := "airdrop"
	validatorAddress := "stridevaloperYYY"
	msg := types.NewMsgClaimAndStake(addr, airdropId, validatorAddress)
	res := msg.GetSignBytes()

	expected := `{"type":"airdrop/MsgClaimAndStake","value":{"airdrop_id":"airdrop","claimer":"strideXXX","validator_address":"stridevaloperYYY"}}`
	require.Equal(t, expected, string(res))
}
