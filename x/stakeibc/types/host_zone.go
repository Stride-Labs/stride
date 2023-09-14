package types

const (
	MaxUnbondingEntries = 7
)

// Per an SDK constraint, we can issue no more than 7 undelegation messages
//   in a given unbonding period
// The unbonding period dictates the cadence (in number of days) with which we submit
//   undelegation messages, such that the 7 messages are spaced out throughout the period
// We calculate this by dividing the period by 7 and then adding 1 as a buffer
// Ex: If our unbonding period is 21 days, we issue an undelegation every 4th day
func (h HostZone) GetUnbondingFrequency() uint64 {
	return (h.UnbondingPeriod / MaxUnbondingEntries) + 1
}
