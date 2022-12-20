package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckExceedsQuota checks if new in/out flow is going to reach the max in/out or not
func (q *Quota) CheckExceedsQuota(direction PacketDirection, amount sdk.Int, totalValue sdk.Int) bool {
	// If there's no channel value (this should be almost impossible), it means there is no
	// supply of the asset, so we shoudn't prevent inflows/outflows
	if totalValue.IsZero() {
		return false
	}
	var threshold sdk.Int
	if direction == PACKET_RECV {
		threshold = totalValue.Mul(q.MaxPercentRecv).Quo(sdk.NewInt(100))
	} else {
		threshold = totalValue.Mul(q.MaxPercentSend).Quo(sdk.NewInt(100))
	}

	return amount.GT(threshold)
}
