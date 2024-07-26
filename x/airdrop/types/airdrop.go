package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns the date index in the allocations array using the current block time
// periodLengthSeconds is the time between each element in the allocations array
func (a *Airdrop) GetCurrentDateIndex(ctx sdk.Context, periodLengthSeconds int64) (dateIndex int, err error) {
	if a.DistributionStartDate == nil {
		return 0, errors.New("distribution start date not set")
	}

	startTime := a.DistributionStartDate.Unix()
	endTime := a.ClawbackDate.Unix()
	blockTime := ctx.BlockTime().Unix()

	if startTime > blockTime {
		return 0, ErrAirdropNotStarted
	}
	if blockTime >= endTime {
		return 0, ErrAirdropEnded
	}

	elapsedTimeSeconds := blockTime - startTime
	elapsedDays := elapsedTimeSeconds / periodLengthSeconds

	// Cap the airdrop index at the last day
	if elapsedDays >= a.GetAirdropPeriods(periodLengthSeconds) {
		elapsedDays = a.GetAirdropPeriods(periodLengthSeconds) - 1
	}

	return int(elapsedDays), nil
}

// Returns number of periods in the airdrop
// periodLengthSeconds is the time between each element in the allocations array
func (a *Airdrop) GetAirdropPeriods(periodLengthSeconds int64) int64 {
	airdropLengthSeconds := int64(a.DistributionEndDate.Unix() - a.DistributionStartDate.Unix())
	numberOfDays := (airdropLengthSeconds / (periodLengthSeconds)) + 1
	return numberOfDays
}
