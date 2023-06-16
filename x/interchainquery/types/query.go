package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Check if a query has timed-out by checking whether the block time is after
// the timeout timestamp
func (q Query) HasTimedOut(ctx sdk.Context) bool {
	return q.TimeoutTimestamp < uint64(ctx.BlockTime().UnixNano())
}
