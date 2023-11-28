package types

import (
	fmt "fmt"
	time "time"
)

// Check if a query has timed-out by checking whether the block time is after
// the timeout timestamp
func (q Query) HasTimedOut(currentBlockTime time.Time) bool {
	return q.TimeoutTimestamp < uint64(currentBlockTime.UnixNano())
}

// Prints an abbreviated query description for logging purposes
func (q Query) Description() string {
	return fmt.Sprintf("QueryId: %s, QueryType: %s, ConnectionId: %s, QueryRequest: %v",
		q.Id, q.QueryType, q.ConnectionId, q.RequestData)
}
