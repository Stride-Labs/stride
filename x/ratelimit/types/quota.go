package types

// CheckExceedsQuota checks if new in/out flow is going to reach the max in/out or not
func (q *Quota) CheckExceedsQuota(direction PacketDirection, amount uint64, totalValue uint64) bool {
	// If there's no channel value (this should be almost impossible), it means there is no
	// supply of the asset, so we shoudn't prevent inflows/outflows
	if totalValue == 0 {
		return false
	}
	if direction == PACKET_RECV {
		return amount > totalValue*q.MaxPercentRecv/100
	}

	return amount > totalValue*q.MaxPercentSend/100
}
