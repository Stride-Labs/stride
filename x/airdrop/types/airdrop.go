package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (a *Airdrop) GetCurrentDateIndex(ctx sdk.Context) (dateIndex int, err error) {
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
