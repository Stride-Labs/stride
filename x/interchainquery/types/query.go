package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Check if a query has timed-out by checking whether the block time is after
// the timeout timestamp
func (q Query) HasTimedOut(ctx sdk.Context) bool {
	return q.TimeoutTimestamp < uint64(ctx.BlockTime().UnixNano())
}

// Prints an abbreviated query description for logging purposes
func (q Query) Description() string {
	return fmt.Sprintf("QueryId: %s, QueryType: %s, ConnectionId: %s, QueryRequest: %v",
		q.Id, q.QueryType, q.ConnectionId, q.RequestData)
}
