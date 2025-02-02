package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v25/app/apptesting"
	"github.com/Stride-Labs/stride/v25/x/airdrop/types"
)

func TestAirdropConfigValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()

	validAirdropId := "airdrop-1"
	validRewardDenom := "denom"
	validDistributorAddress := validNonAdminAddress
	validAllocatorAddress := validNonAdminAddress
	validLinkerAddress := validNonAdminAddress

	validDistributionStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validDistributionEndDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	validClawbackDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	validDeadlineDate := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	startDateMinusDelta := validDistributionStartDate.Add(-1 * time.Hour)
	endDatePlusDelta := validDistributionEndDate.Add(time.Hour)
	endDateMinusDelta := validDistributionEndDate.Add(-1 * time.Hour)

	validEarlyClaimPenalty := sdk.MustNewDecFromStr("0.5")

	testCases := []struct {
		name                  string
		airdropId             string
		rewardDenom           string
		distributionStartDate *time.Time
		distributionEndDate   *time.Time
		clawbackDate          *time.Time
		claimTypeDeadlineDate *time.Time
		earlyClaimPenalty     sdk.Dec
		distributorAddress    string
		allocatorAddress      string
		linkerAddress         string
		expectedError         string
	}{
		{
			name:                  "valid message",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
		},
		{
			name:                  "invalid airdrop id",
			airdropId:             "",
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "airdrop-id must be specified",
		},
		{
			name:                  "invalid reward denom",
			airdropId:             validAirdropId,
			rewardDenom:           "",
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "reward denom must be specified",
		},
		{
			name:                  "nil distribution start date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution start date must be specified",
		},
		{
			name:                  "nil distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution end date must be specified",
		},
		{
			name:                  "nil clawback date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "clawback date must be specified",
		},
		{
			name:                  "nil deadline date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "deadline date must be specified",
		},

		{
			name:                  "empty distribution start date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &(time.Time{}),
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution start date must be specified",
		},
		{
			name:                  "empty distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &(time.Time{}),
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution end date must be specified",
		},
		{
			name:                  "empty clawback date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &(time.Time{}),
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "clawback date must be specified",
		},
		{
			name:                  "empty deadline date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &(time.Time{}),
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "deadline date must be specified",
		},
		{
			name:                  "distribution start date equals distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionStartDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution end date must be after the start date",
		},
		{
			name:                  "distribution start date after distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &endDatePlusDelta,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "distribution end date must be after the start date",
		},
		{
			name:                  "claim type deadline date equals distribution start date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDistributionStartDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "claim type deadline date must be after the distribution start date",
		},
		{
			name:                  "claim type deadline date before distribution start date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &startDateMinusDelta,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "claim type deadline date must be after the distribution start date",
		},
		{
			name:                  "claim type deadline date equal distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDistributionEndDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "claim type deadline date must be before the distribution end date",
		},
		{
			name:                  "claim type deadline date after distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &endDatePlusDelta,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "claim type deadline date must be before the distribution end date",
		},
		{
			name:                  "clawback date equals distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validDistributionEndDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "clawback date must be after the distribution end date",
		},
		{
			name:                  "clawback date before distribution end date",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &endDateMinusDelta,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "clawback date must be after the distribution end date",
		},
		{
			name:                  "nil early claim penalty",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "early claim penalty must be specified",
		},
		{
			name:                  "early claim penalty less than 0",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     sdk.NewDec(-1),
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "early claim penalty must be between 0 and 1",
		},
		{
			name:                  "early claim penalty less than 0",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     sdk.NewDec(-1),
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "early claim penalty must be between 0 and 1",
		},
		{
			name:                  "invalid distributor address",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    invalidAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "invalid distributor address",
		},
		{
			name:                  "invalid allocator address",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      invalidAddress,
			linkerAddress:         validLinkerAddress,
			expectedError:         "invalid allocator address",
		},
		{
			name:                  "invalid linker address",
			airdropId:             validAirdropId,
			rewardDenom:           validRewardDenom,
			distributionStartDate: &validDistributionStartDate,
			distributionEndDate:   &validDistributionEndDate,
			clawbackDate:          &validClawbackDate,
			claimTypeDeadlineDate: &validDeadlineDate,
			earlyClaimPenalty:     validEarlyClaimPenalty,
			distributorAddress:    validDistributorAddress,
			allocatorAddress:      validAllocatorAddress,
			linkerAddress:         invalidAddress,
			expectedError:         "invalid linker address",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualError := types.AirdropConfigValidateBasic(
				tc.airdropId,
				tc.rewardDenom,
				tc.distributionStartDate,
				tc.distributionEndDate,
				tc.clawbackDate,
				tc.claimTypeDeadlineDate,
				tc.earlyClaimPenalty,
				tc.distributorAddress,
				tc.allocatorAddress,
				tc.linkerAddress,
			)
			if tc.expectedError != "" {
				require.ErrorContains(t, actualError, tc.expectedError)
				return
			}
			require.NoError(t, actualError)
		})
	}
}
