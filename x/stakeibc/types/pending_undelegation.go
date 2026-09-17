package types

import sdkmath "cosmossdk.io/math"

// PendingUndelegation is a one-shot undelegation amount queued for a host zone (e.g. by an
// upgrade handler) that is submitted through the normal undelegate pipeline at the next day epoch
//
// This is intentionally not a proto type: it is never exported over gRPC or genesis, and the
// store value is just the marshalled sdkmath.Int
type PendingUndelegation struct {
	ChainId string
	Amount  sdkmath.Int
}
