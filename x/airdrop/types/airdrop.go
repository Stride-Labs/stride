package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns the date index in the allocations array using the current block time
func (a *Airdrop) GetCurrentDateIndex(ctx sdk.Context) (dateIndex int, err error) {
	if a.DistributionStartDate == nil {
		return 0, errors.New("distribution start date not set")
	}
	if a.DistributionEndDate == nil {
		return 0, errors.New("distribution end date not set")
	}

	startTime := a.DistributionStartDate.Unix()
	endTime := a.DistributionEndDate.Unix()
	blockTime := ctx.BlockTime().Unix()

	if startTime > blockTime {
		return 0, ErrDistributionNotStarted
	}
	if blockTime > endTime {
		return 0, ErrDistributionEnded
	}

	elapsedTimeSeconds := blockTime - startTime
	elapsedDays := elapsedTimeSeconds / (60 * 60 * 24)

	return int(elapsedDays), nil
}
