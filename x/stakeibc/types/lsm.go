package types

import (
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
)

// Maintains the context and progress of an LSM Liquid Stake in the event
// that the transaction finishes asynchonously after the validator exchange rate query
type LSMLiquidStake struct {
	Deposit   recordstypes.LSMTokenDeposit `json:"deposit"`
	HostZone  HostZone                     `json:"host_zone"`
	Validator Validator                    `json:"validator"`
}
