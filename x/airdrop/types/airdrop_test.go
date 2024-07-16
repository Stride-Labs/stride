package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func TestGetCurrentDateIndex(t *testing.T) {
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(time.Hour * 24 * 150) // 150 days later
	windowLength := int64(time.Hour * 24)

	airdrop := types.Airdrop{
		DistributionStartDate: &startTime,
		DistributionEndDate:   &endTime,
	}

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
			currentTime:       startTime.Add((time.Hour * 23) + (time.Minute * 59) + (time.Second * 59)),
			expectedDateIndex: 0,
		},
		{
			name:              "start of second day",
			currentTime:       startTime.Add(time.Hour * 24),
			expectedDateIndex: 1,
		},
		{
			name:              "start of third day",
			currentTime:       startTime.Add(time.Hour * 48),
			expectedDateIndex: 2,
		},
		{
			name:              "middle of third day",
			currentTime:       startTime.Add(time.Hour * 60),
			expectedDateIndex: 2,
		},
		{
			name:              "100 days later",
			currentTime:       startTime.Add(time.Hour * 24 * 100),
			expectedDateIndex: 99,
		},
		{
			name:          "airdrop not started",
			currentTime:   startTime.Add(-1 * time.Minute),
			expectedError: types.ErrDistributionNotStarted,
		},
		{
			name:          "airdrop already ended",
			currentTime:   endTime.Add(time.Minute),
			expectedError: types.ErrDistributionEnded,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}
			ctx = ctx.WithBlockTime(tc.currentTime)

			actualDateIndex, actualError := airdrop.GetCurrentDateIndex(ctx, windowLength)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedDateIndex, actualDateIndex, "date index")
			} else {
				require.ErrorIs(t, tc.expectedError, actualError)
			}
		})
	}
}
