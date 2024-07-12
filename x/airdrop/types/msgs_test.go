package types_test

import (
	"regexp"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// ----------------------------------------------
//               MsgCreateAirdrop
// ----------------------------------------------

func TestMsgCreateAirdrop_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAirdropId := "airdrop-1"
	validRewardDenom := "denom"
	validDistributionAddress := validNonAdminAddress

	validDistributionStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validDistributionEndDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	validClawbackDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	validDeadlineDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	startDateMinusDelta := validDistributionStartDate.Add(-1 * time.Hour)
	endDatePlusDelta := validDistributionEndDate.Add(time.Hour)
	endDateMinusDelta := validDistributionEndDate.Add(-1 * time.Hour)

	validEarlyClaimPenalty := sdk.MustNewDecFromStr("0.5")
	validClaimAndStakeBonus := sdk.MustNewDecFromStr("0.05")

	tests := []struct {
		name          string
		msg           types.MsgCreateAirdrop
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
		},
		{
			name: "invalid admin address",
			msg: types.MsgCreateAirdrop{
				Admin:                 invalidAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "invalid address",
		},
		{
			name: "not admin address",
			msg: types.MsgCreateAirdrop{
				Admin:                 validNonAdminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "is not an admin",
		},
		{
			name: "invalid airdrop id",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             "",
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "invalid reward denom",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           "",
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "reward denom must be specified",
		},
		{
			name: "nil distribution start date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution start date must be specified",
		},
		{
			name: "nil distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be specified",
		},
		{
			name: "nil clawback date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be specified",
		},
		{
			name: "nil deadline date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "deadline date must be specified",
		},

		{
			name: "empty distribution start date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &(time.Time{}),
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution start date must be specified",
		},
		{
			name: "empty distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &(time.Time{}),
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be specified",
		},
		{
			name: "empty clawback date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &(time.Time{}),
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be specified",
		},
		{
			name: "empty deadline date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &(time.Time{}),
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "deadline date must be specified",
		},

		{
			name: "distribution start date equals distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionStartDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be after the start date",
		},
		{
			name: "distribution start date after distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &endDatePlusDelta,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be after the start date",
		},
		{
			name: "claim type deadline date equals distribution start date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDistributionStartDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be after the distribution start date",
		},
		{
			name: "claim type deadline date before distribution start date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &startDateMinusDelta,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be after the distribution start date",
		},
		{
			name: "claim type deadline date equal distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDistributionEndDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be before the distribution end date",
		},
		{
			name: "claim type deadline date after distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &endDatePlusDelta,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be before the distribution end date",
		},
		{
			name: "clawback date equals distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validDistributionEndDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be after the distribution end date",
		},
		{
			name: "clawback date before distribution end date",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &endDateMinusDelta,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be after the distribution end date",
		},
		{
			name: "nil early claim penalty",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be specified",
		},
		{
			name: "nil claim and stake bonus",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim and stake bonus must be specified",
		},
		{
			name: "early claim penalty less than 0",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     sdk.NewDec(-1),
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "early claim penalty less than 0",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     sdk.NewDec(-1),
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "claim and stake bonus greater than 1",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    sdk.NewDec(2),
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "claim and stake bonus greater than 1",
			msg: types.MsgCreateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    sdk.NewDec(2),
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
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

func TestMsgCreateAirdrop_GetSignBytes(t *testing.T) {
	admin := "admin"
	airdropId := "airdrop"
	distributionAddress := "distributor"
	rewardDenom := "denom"

	distributionStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	distributionEndDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clawbackDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	deadlineDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	earlyClaimPenalty := sdk.MustNewDecFromStr("0.5")
	claimAndStakeBonus := sdk.MustNewDecFromStr("0.05")

	msg := types.NewMsgCreateAirdrop(
		admin,
		airdropId,
		rewardDenom,
		&distributionStartDate,
		&distributionEndDate,
		&clawbackDate,
		&deadlineDate,
		earlyClaimPenalty,
		claimAndStakeBonus,
		distributionAddress,
	)
	res := msg.GetSignBytes()

	expected := strings.TrimSpace(`
		{"type":"airdrop/MsgCreateAirdrop",
		"value":{"admin":"admin",
		"airdrop_id":"airdrop",
		"claim_and_stake_bonus":"0.050000000000000000",
		"claim_type_deadline_date":"2024-02-01T00:00:00Z",
		"clawback_date":"2024-07-01T00:00:00Z",
		"distribution_address":"distributor",
		"distribution_end_date":"2024-06-01T00:00:00Z",
		"distribution_start_date":"2024-01-01T00:00:00Z",
		"early_claim_penalty":"0.500000000000000000",
		"reward_denom":"denom"}}`)

	re := regexp.MustCompile(`\s+`)
	expected = re.ReplaceAllString(expected, "")
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgUpdateAirdrop
// ----------------------------------------------

func TestMsgUpdateAirdrop_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAirdropId := "airdrop-1"
	validRewardDenom := "denom"
	validDistributionAddress := validNonAdminAddress

	validDistributionStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validDistributionEndDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	validClawbackDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	validDeadlineDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	startDateMinusDelta := validDistributionStartDate.Add(-1 * time.Hour)
	endDatePlusDelta := validDistributionEndDate.Add(time.Hour)
	endDateMinusDelta := validDistributionEndDate.Add(-1 * time.Hour)

	validEarlyClaimPenalty := sdk.MustNewDecFromStr("0.5")
	validClaimAndStakeBonus := sdk.MustNewDecFromStr("0.05")

	tests := []struct {
		name          string
		msg           types.MsgUpdateAirdrop
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
		},
		{
			name: "invalid admin address",
			msg: types.MsgUpdateAirdrop{
				Admin:                 invalidAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "invalid address",
		},
		{
			name: "not admin address",
			msg: types.MsgUpdateAirdrop{
				Admin:                 validNonAdminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "is not an admin",
		},
		{
			name: "invalid airdrop id",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             "",
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "invalid reward denom",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           "",
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "reward denom must be specified",
		},
		{
			name: "nil distribution start date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution start date must be specified",
		},
		{
			name: "nil distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be specified",
		},
		{
			name: "nil clawback date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be specified",
		},
		{
			name: "nil deadline date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "deadline date must be specified",
		},

		{
			name: "empty distribution start date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &(time.Time{}),
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution start date must be specified",
		},
		{
			name: "empty distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &(time.Time{}),
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be specified",
		},
		{
			name: "empty clawback date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &(time.Time{}),
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be specified",
		},
		{
			name: "empty deadline date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &(time.Time{}),
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "deadline date must be specified",
		},

		{
			name: "distribution start date equals distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionStartDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be after the start date",
		},
		{
			name: "distribution start date after distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &endDatePlusDelta,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "distribution end date must be after the start date",
		},
		{
			name: "claim type deadline date equals distribution start date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDistributionStartDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be after the distribution start date",
		},
		{
			name: "claim type deadline date before distribution start date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &startDateMinusDelta,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be after the distribution start date",
		},
		{
			name: "claim type deadline date equal distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDistributionEndDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be before the distribution end date",
		},
		{
			name: "claim type deadline date after distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &endDatePlusDelta,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim type deadline date must be before the distribution end date",
		},
		{
			name: "clawback date equals distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validDistributionEndDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be after the distribution end date",
		},
		{
			name: "clawback date before distribution end date",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &endDateMinusDelta,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "clawback date must be after the distribution end date",
		},
		{
			name: "nil early claim penalty",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be specified",
		},
		{
			name: "nil claim and stake bonus",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "claim and stake bonus must be specified",
		},
		{
			name: "early claim penalty less than 0",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     sdk.NewDec(-1),
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "early claim penalty less than 0",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     sdk.NewDec(-1),
				ClaimAndStakeBonus:    validClaimAndStakeBonus,
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "claim and stake bonus greater than 1",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    sdk.NewDec(2),
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
		},
		{
			name: "claim and stake bonus greater than 1",
			msg: types.MsgUpdateAirdrop{
				Admin:                 adminAddress,
				AirdropId:             validAirdropId,
				RewardDenom:           validRewardDenom,
				DistributionStartDate: &validDistributionStartDate,
				DistributionEndDate:   &validDistributionEndDate,
				ClawbackDate:          &validClawbackDate,
				ClaimTypeDeadlineDate: &validDeadlineDate,
				EarlyClaimPenalty:     validEarlyClaimPenalty,
				ClaimAndStakeBonus:    sdk.NewDec(2),
				DistributionAddress:   validDistributionAddress,
			},
			expectedError: "early claim penalty must be between 0 and 1",
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

func TestMsgUpdateAirdrop_GetSignBytes(t *testing.T) {
	admin := "admin"
	airdropId := "airdrop"
	distributionAddress := "distributor"
	rewardDenom := "denom"

	distributionStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	distributionEndDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clawbackDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	deadlineDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	earlyClaimPenalty := sdk.MustNewDecFromStr("0.5")
	claimAndStakeBonus := sdk.MustNewDecFromStr("0.05")

	msg := types.NewMsgUpdateAirdrop(
		admin,
		airdropId,
		rewardDenom,
		&distributionStartDate,
		&distributionEndDate,
		&clawbackDate,
		&deadlineDate,
		earlyClaimPenalty,
		claimAndStakeBonus,
		distributionAddress,
	)
	res := msg.GetSignBytes()

	expected := strings.TrimSpace(`
		{"type":"airdrop/MsgUpdateAirdrop",
		"value":{"admin":"admin",
		"airdrop_id":"airdrop",
		"claim_and_stake_bonus":"0.050000000000000000",
		"claim_type_deadline_date":"2024-02-01T00:00:00Z",
		"clawback_date":"2024-07-01T00:00:00Z",
		"distribution_address":"distributor",
		"distribution_end_date":"2024-06-01T00:00:00Z",
		"distribution_start_date":"2024-01-01T00:00:00Z",
		"early_claim_penalty":"0.500000000000000000",
		"reward_denom":"denom"}}`)

	re := regexp.MustCompile(`\s+`)
	expected = re.ReplaceAllString(expected, "")
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgAddAllocations
// ----------------------------------------------

func TestMsgAddAllocations_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAirdropId := "airdrop-1"
	validAllocations := []types.RawAllocation{
		{
			UserAddress: "user-1",
			Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
		},
		{
			UserAddress: "user-2",
			Allocations: []sdkmath.Int{sdkmath.NewInt(1)},
		},
		{
			UserAddress: "user-3",
			Allocations: []sdkmath.Int{sdkmath.NewInt(2)},
		},
	}

	tests := []struct {
		name          string
		msg           types.MsgAddAllocations
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgAddAllocations{
				Admin:       adminAddress,
				AirdropId:   validAirdropId,
				Allocations: validAllocations,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgAddAllocations{
				Admin:       invalidAddress,
				AirdropId:   validAirdropId,
				Allocations: validAllocations,
			},
			expectedError: "invalid address",
		},
		{
			name: "not admin address",
			msg: types.MsgAddAllocations{
				Admin:       validNonAdminAddress,
				AirdropId:   validAirdropId,
				Allocations: validAllocations,
			},
			expectedError: "is not an admin",
		},
		{
			name: "missing airdrop id",
			msg: types.MsgAddAllocations{
				Admin:       adminAddress,
				AirdropId:   "",
				Allocations: validAllocations,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "missing address",
			msg: types.MsgAddAllocations{
				Admin:     adminAddress,
				AirdropId: validAirdropId,
				Allocations: []types.RawAllocation{
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
					},
					{
						UserAddress: "",
						Allocations: []sdkmath.Int{sdkmath.NewInt(1)},
					},
				},
			},
			expectedError: "all addresses in allocations must be specified",
		},
		{
			name: "nil allocation",
			msg: types.MsgAddAllocations{
				Admin:     adminAddress,
				AirdropId: validAirdropId,
				Allocations: []types.RawAllocation{
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{{}},
					},
					{
						UserAddress: "user-2",
						Allocations: []sdkmath.Int{sdkmath.NewInt(-1)},
					},
				},
			},
			expectedError: "all allocation amounts must be specified and positive",
		},
		{
			name: "negative allocation",
			msg: types.MsgAddAllocations{
				Admin:     adminAddress,
				AirdropId: validAirdropId,
				Allocations: []types.RawAllocation{
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
					},
					{
						UserAddress: "user-2",
						Allocations: []sdkmath.Int{sdkmath.NewInt(-1)},
					},
				},
			},
			expectedError: "all allocation amounts must be specified and positive",
		},
		{
			name: "duplicate address",
			msg: types.MsgAddAllocations{
				Admin:     adminAddress,
				AirdropId: validAirdropId,
				Allocations: []types.RawAllocation{
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
					},
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{sdkmath.NewInt(-1)},
					},
				},
			},
			expectedError: "address user-1 is specified more than once",
		},
		{
			name: "inconsistent allocations length",
			msg: types.MsgAddAllocations{
				Admin:     adminAddress,
				AirdropId: validAirdropId,
				Allocations: []types.RawAllocation{
					{
						UserAddress: "user-1",
						Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
					},
					{
						UserAddress: "user-2",
						Allocations: []sdkmath.Int{sdkmath.NewInt(1)},
					},
					{
						UserAddress: "user-3",
						Allocations: []sdkmath.Int{sdkmath.NewInt(1), sdkmath.NewInt(2)},
					},
				},
			},
			expectedError: "address user-3 has an inconsistent number of allocations",
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

func TestMsgAddAllocations_GetSignBytes(t *testing.T) {
	admin := "admin"
	airdropId := "airdrop"
	allocations := []types.RawAllocation{
		{
			UserAddress: "user-1",
			Allocations: []sdkmath.Int{sdkmath.NewInt(0)},
		},
		{
			UserAddress: "user-2",
			Allocations: []sdkmath.Int{sdkmath.NewInt(1)},
		},
	}

	msg := types.NewMsgAddAllocations(admin, airdropId, allocations)
	res := msg.GetSignBytes()

	expected := strings.TrimSpace(`
		{"type":"airdrop/MsgAddAllocations",
		"value":{"admin":"admin",
		"airdrop_id":"airdrop",
		"allocations":[{"allocations":["0"],"user_address":"user-1"},{"allocations":["1"],"user_address":"user-2"}]}}`)

	re := regexp.MustCompile(`\s+`)
	expected = re.ReplaceAllString(expected, "")
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgUpdateUserAllocation
// ----------------------------------------------

func TestMsgUpdateUserAllocation_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAirdropId := "airdrop-1"
	validUser := "user"
	validAllocation := []sdkmath.Int{sdkmath.NewInt(0), sdkmath.NewInt(1)}

	tests := []struct {
		name          string
		msg           types.MsgUpdateUserAllocation
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgUpdateUserAllocation{
				Admin:       adminAddress,
				AirdropId:   validAirdropId,
				UserAddress: validUser,
				Allocations: validAllocation,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgUpdateUserAllocation{
				Admin:       invalidAddress,
				AirdropId:   validAirdropId,
				UserAddress: validUser,
				Allocations: validAllocation,
			},
			expectedError: "invalid address",
		},
		{
			name: "not admin address",
			msg: types.MsgUpdateUserAllocation{
				Admin:       validNonAdminAddress,
				AirdropId:   validAirdropId,
				UserAddress: validUser,
				Allocations: validAllocation,
			},
			expectedError: "is not an admin",
		},
		{
			name: "missing airdrop id",
			msg: types.MsgUpdateUserAllocation{
				Admin:       adminAddress,
				AirdropId:   "",
				UserAddress: validUser,
				Allocations: validAllocation,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "missing address",
			msg: types.MsgUpdateUserAllocation{
				Admin:       adminAddress,
				AirdropId:   validAirdropId,
				UserAddress: "",
				Allocations: validAllocation,
			},
			expectedError: "user address must be specified",
		},
		{
			name: "nil allocation",
			msg: types.MsgUpdateUserAllocation{
				Admin:       adminAddress,
				AirdropId:   validAirdropId,
				UserAddress: validUser,
				Allocations: []sdkmath.Int{{}},
			},
			expectedError: "all allocation amounts must be specified and positive",
		},
		{
			name: "negative allocation",
			msg: types.MsgUpdateUserAllocation{
				Admin:       adminAddress,
				AirdropId:   validAirdropId,
				UserAddress: validUser,
				Allocations: []sdkmath.Int{sdkmath.NewInt(-1)},
			},
			expectedError: "all allocation amounts must be specified and positive",
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

func TestMsgUpdateUserAllocation_GetSignBytes(t *testing.T) {
	admin := "admin"
	airdropId := "airdrop"
	userAddress := "user"
	allocation := []sdkmath.Int{sdkmath.NewInt(0)}

	msg := types.NewMsgUpdateUserAllocation(admin, airdropId, userAddress, allocation)
	res := msg.GetSignBytes()

	expected := strings.TrimSpace(`
		{"type":"airdrop/MsgUpdateUserAllocation",
		"value":{"admin":"admin",
		"airdrop_id":"airdrop",
		"allocations":["0"],
		"user_address": "user"}}`)

	re := regexp.MustCompile(`\s+`)
	expected = re.ReplaceAllString(expected, "")
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgLinkAddresses
// ----------------------------------------------

func TestMsgLinkAddresses_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAirdropId := "airdrop-1"
	validStrideAddress := validNonAdminAddress
	validHostAddress := "hostXXX"

	tests := []struct {
		name          string
		msg           types.MsgLinkAddresses
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgLinkAddresses{
				Admin:         adminAddress,
				AirdropId:     validAirdropId,
				StrideAddress: validStrideAddress,
				HostAddress:   validHostAddress,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgLinkAddresses{
				Admin:         invalidAddress,
				AirdropId:     validAirdropId,
				StrideAddress: validStrideAddress,
				HostAddress:   validHostAddress,
			},
			expectedError: "invalid address",
		},
		{
			name: "not admin address",
			msg: types.MsgLinkAddresses{
				Admin:         validNonAdminAddress,
				AirdropId:     validAirdropId,
				StrideAddress: validStrideAddress,
				HostAddress:   validHostAddress,
			},
			expectedError: "is not an admin",
		},
		{
			name: "missing airdrop id",
			msg: types.MsgLinkAddresses{
				Admin:         adminAddress,
				AirdropId:     "",
				StrideAddress: validStrideAddress,
				HostAddress:   validHostAddress,
			},
			expectedError: "airdrop-id must be specified",
		},
		{
			name: "invalid stride address",
			msg: types.MsgLinkAddresses{
				Admin:         adminAddress,
				AirdropId:     validAirdropId,
				StrideAddress: invalidAddress,
				HostAddress:   validHostAddress,
			},
			expectedError: "invalid stride address",
		},
		{
			name: "invalid host address",
			msg: types.MsgLinkAddresses{
				Admin:         adminAddress,
				AirdropId:     validAirdropId,
				StrideAddress: validStrideAddress,
				HostAddress:   "",
			},
			expectedError: "host address must be specified",
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

func TestMsgLinkAddresses_GetSignBytes(t *testing.T) {
	admin := "admin"
	airdropId := "airdrop"
	strideAddress := "stride"
	hostAddress := "host"

	msg := types.NewMsgLinkAddresses(admin, airdropId, strideAddress, hostAddress)
	res := msg.GetSignBytes()

	expected := strings.TrimSpace(`
		{"type":"airdrop/MsgLinkAddresses",
		"value":{"admin":"admin",
		"airdrop_id":"airdrop",
		"host_address":"host",
		"stride_address": "stride"}}`)

	re := regexp.MustCompile(`\s+`)
	expected = re.ReplaceAllString(expected, "")
	require.Equal(t, expected, string(res))
}
