package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// Helper function to cast an array of allocations as int64's into sdkmath.Ints
func allocationsToSdkInt(allocationsInt64 []int64) (allocationsSdkInt []sdkmath.Int) {
	for _, allocation := range allocationsInt64 {
		allocationsSdkInt = append(allocationsSdkInt, sdkmath.NewInt(allocation))
	}
	return allocationsSdkInt
}

// Helper function to cast an array of allocations as sdk.Int's into int64
func allocationsToInt64(allocationsSdkInt []sdkmath.Int) (allocationsInt64 []int64) {
	for _, allocation := range allocationsSdkInt {
		allocationsInt64 = append(allocationsInt64, allocation.Int64())
	}
	return allocationsInt64
}

func (s *KeeperTestSuite) TestClaimDaily() {
	testCases := []struct {
		name                string
		timeOffset          time.Duration
		initialAllocations  []int64
		expectedAllocations []int64
		initialClaimed      int64
		expectedClaimed     int64
		expectedNewRewards  int64
		initialClaimType    types.ClaimType
		expectedError       string
	}{
		{
			name:                "claim the first day",
			timeOffset:          time.Hour, // one hour into first window
			initialAllocations:  []int64{10, 10, 10},
			expectedAllocations: []int64{0, 10, 10},
			initialClaimed:      100,
			expectedClaimed:     100 + 10,
			expectedNewRewards:  10,
			initialClaimType:    types.UNSPECIFIED,
		},
		{
			name:                "claim the second day",
			timeOffset:          time.Hour * 25, // one hour into second window
			initialAllocations:  []int64{10, 10, 10},
			expectedAllocations: []int64{0, 0, 10},
			initialClaimed:      100,
			expectedClaimed:     100 + 20,
			expectedNewRewards:  20,
			initialClaimType:    types.UNSPECIFIED,
		},
		{
			name:                "claim all days at once",
			timeOffset:          time.Hour * 49, // one hour into third window
			initialAllocations:  []int64{10, 10, 10},
			expectedAllocations: []int64{0, 0, 0},
			initialClaimed:      100,
			expectedClaimed:     100 + 30,
			expectedNewRewards:  30,
			initialClaimType:    types.UNSPECIFIED,
		},
		{
			name:                "claim with previous claims",
			timeOffset:          time.Hour * 49, // one hour into third window
			initialAllocations:  []int64{0, 0, 10, 20},
			expectedAllocations: []int64{0, 0, 0, 20},
			initialClaimed:      100,
			expectedClaimed:     100 + 10,
			expectedNewRewards:  10,
			initialClaimType:    types.DAILY,
		},
		{
			name:               "already chose to claim early",
			timeOffset:         time.Hour,
			initialAllocations: []int64{},
			initialClaimed:     100,
			initialClaimType:   types.CLAIM_EARLY,
			expectedError:      "user has already elected claim option",
		},
		{
			name:               "already chose to claim and stake",
			timeOffset:         time.Hour,
			initialAllocations: []int64{},
			initialClaimed:     100,
			initialClaimType:   types.CLAIM_AND_STAKE,
			expectedError:      "user has already elected claim option",
		},
		{
			name:               "no rewards today",
			timeOffset:         time.Hour * 49, // one hour into third window
			initialAllocations: []int64{0, 0, 0, 20},
			initialClaimed:     100,
			initialClaimType:   types.UNSPECIFIED,
			expectedError:      "no unclaimed rewards for today",
		},
		{
			name:               "airdrop not started yet",
			timeOffset:         (-1 * time.Hour), // before airdrop start
			initialAllocations: []int64{},
			initialClaimed:     100,
			initialClaimType:   types.UNSPECIFIED,
			expectedError:      "airdrop distribution has not started",
		},
		{
			name:               "airdrop ended",
			timeOffset:         time.Hour * 24 * 1000, // far into future
			initialAllocations: []int64{},
			initialClaimed:     100,
			initialClaimType:   types.UNSPECIFIED,
			expectedError:      "airdrop distribution has ended",
		},
		{
			name:               "not enough distributor funds",
			timeOffset:         time.Hour, // far into future
			initialAllocations: []int64{1000000},
			initialClaimed:     100,
			initialClaimType:   types.UNSPECIFIED,
			expectedError:      "unable to distribute rewards",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset state

			claimer := s.TestAccs[0]
			distributor := s.TestAccs[1]

			// Fund the distributor
			initialDistributorBalance := sdk.NewInt(1000)
			s.FundAccount(distributor, sdk.NewCoin(RewardDenom, initialDistributorBalance))

			// Create the initial airdrop config
			s.App.AirdropKeeper.SetAirdrop(s.Ctx, types.Airdrop{
				Id:                    AirdropId,
				RewardDenom:           RewardDenom,
				DistributionAddress:   distributor.String(),
				DistributionStartDate: &DistributionStartDate,
				DistributionEndDate:   &DistributionEndDate,
			})

			// Set the block time to the distribution start time plus the offset
			blockTime := DistributionStartDate.Add(tc.timeOffset)
			s.Ctx = s.Ctx.WithBlockTime(blockTime)

			// Create the initial user and allocations
			s.App.AirdropKeeper.SetUserAllocation(s.Ctx, types.UserAllocation{
				AirdropId:   AirdropId,
				Address:     claimer.String(),
				ClaimType:   tc.initialClaimType,
				Claimed:     sdkmath.NewInt(tc.initialClaimed),
				Allocations: allocationsToSdkInt(tc.initialAllocations),
			})

			// Call claim daily
			actualError := s.App.AirdropKeeper.ClaimDaily(s.Ctx, AirdropId, claimer.String())
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError)
				return
			}
			s.Require().NoError(actualError, "no error expected when claiming daily")

			// Check that the user was updated
			userAllocation := s.MustGetUserAllocation(AirdropId, claimer.String())
			s.Require().Equal(tc.expectedAllocations, allocationsToInt64(userAllocation.Allocations), "allocations")
			s.Require().Equal(tc.expectedClaimed, userAllocation.Claimed.Int64(), "claimed")
			s.Require().Equal(types.DAILY, userAllocation.ClaimType, "claim types")

			// Confirm funds were decremented from the distributor
			expectedDistributorBalance := initialDistributorBalance.Sub(sdkmath.NewInt(tc.expectedNewRewards))
			actualDistributorBalance := s.App.BankKeeper.GetBalance(s.Ctx, distributor, RewardDenom).Amount
			s.Require().Equal(expectedDistributorBalance.Int64(), actualDistributorBalance.Int64(), "distributor balance")

			// Confirm funds were sent to the user
			claimerBalance := s.App.BankKeeper.GetBalance(s.Ctx, claimer, RewardDenom).Amount
			s.Require().Equal(tc.expectedNewRewards, claimerBalance.Int64(), "claimer balance")
		})
	}
}
