package types

// CheckExceedsQuota checks if new in/out flow is going to reach the max in/out or not
func (q Quota) CheckExceedsQuota(direction PacketDirection, amount uint64, totalValue uint64) bool {
	if direction == PACKET_RECV {
		return amount > totalValue*q.MaxPercentRecv/100
	}

	return amount > totalValue*q.MaxPercentSend/100
}
