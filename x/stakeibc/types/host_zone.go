package types

const (
	MaxUnbondingEntries = 7
)

func (h HostZone) GetTotalDelegations() {
	// TODO [LSM]
}

func (h HostZone) GetUnbondingFrequency() uint64 {
	return (h.UnbondingPeriod / MaxUnbondingEntries) + 1
}
