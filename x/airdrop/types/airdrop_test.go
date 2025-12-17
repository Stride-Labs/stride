package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v31/x/airdrop/types"
)

func TestGetCurrentDateIndex(t *testing.T) {
	// Setup: 10 day long airdrop from 1/1 to 1/10 with clawback on 1/15
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
	clawbackTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	windowLengthSeconds := int64(24 * 60 * 60)

	airdrop := types.Airdrop{
		DistributionStartDate: &startTime,
		DistributionEndDate:   &endTime,
		ClawbackDate:          &clawbackTime,
	}
	require.Equal(t, airdrop.GetAirdropPeriods(windowLengthSeconds), int64(10), "airdrop length in setup")

	testCases := []struct {
		name              string
		currentTime       time.Time
		expectedDateIndex int
		expectedError     error
	}{
		{
			name:              "start time",
			currentTime:       startTime,
			expectedDateIndex: 0,
		},
		{
			name:              "start time plus 1 second",
			currentTime:       startTime.Add(time.Second),
			expectedDateIndex: 0,
		},
		{
			name:              "start time plus 12 hours",
			currentTime:       startTime.Add(time.Hour * 12),
			expectedDateIndex: 0,
		},
		{
			name:              "one second before second day",
			currentTime:       startTime.Add((time.Hour * 24) - time.Second),
			expectedDateIndex: 0,
		},
		{
			name:              "start of second day",
			currentTime:       startTime.Add(time.Hour * 24),
			expectedDateIndex: 1,
		},
		{
			name:              "start of third day",
			currentTime:       startTime.Add(time.Hour * (24 + 24)),
			expectedDateIndex: 2,
		},
		{
			name:              "middle of third day",
			currentTime:       startTime.Add(time.Hour * (24 + 24 + 16)),
			expectedDateIndex: 2,
		},
		{
			name:              "middle of fifth day",
			currentTime:       startTime.Add(time.Hour * ((24 * 4) + 16)),
			expectedDateIndex: 4,
		},
		{
			name:              "last day of distribution",
			currentTime:       endTime,
			expectedDateIndex: 9,
		},
		{
			name:              "last second of last day of distribution",
			currentTime:       endTime.Add((time.Hour * 24) - 1),
			expectedDateIndex: 9,
		},
		{
			name:              "last second before clawback",
			currentTime:       clawbackTime.Add(-1 * time.Second),
			expectedDateIndex: 9, // gets capped at last index
		},
		{
			name:          "airdrop not started",
			currentTime:   startTime.Add(-1 * time.Minute),
			expectedError: types.ErrAirdropNotFound,
		},
		{
			name:          "airdrop already ended",
			currentTime:   clawbackTime,
			expectedError: types.ErrAirdropEnded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}
			ctx = ctx.WithBlockTime(tc.currentTime)

			actualDateIndex, actualError := airdrop.GetCurrentDateIndex(ctx, windowLengthSeconds)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedDateIndex, actualDateIndex, "date index")
			} else {
				require.ErrorIs(t, tc.expectedError, actualError)
			}
		})
	}
}

func TestGetAirdropPeriods(t *testing.T) {
	windowLengthSeconds := int64(24 * 60 * 60)

	testCases := []struct {
		name           string
		startDate      time.Time
		endDate        time.Time
		expectedLength int64
	}{
		{
			name:           "one day",
			startDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedLength: 1,
		},
		{
			name:           "two days",
			startDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expectedLength: 2,
		},
		{
			name:           "five days",
			startDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			expectedLength: 5,
		},
		{
			name:           "one month",
			startDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedLength: 32,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			airdrop := types.Airdrop{
				DistributionStartDate: &tc.startDate,
				DistributionEndDate:   &tc.endDate,
			}
			actualLength := airdrop.GetAirdropPeriods(windowLengthSeconds)
			require.Equal(t, tc.expectedLength, actualLength, "airdrop length")
		})
	}
}
