package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Initializes a new flow from the channel value
func NewFlow(channelValue uint64) Flow {
	flow := Flow{
		ChannelValue: channelValue,
		Inflow:       0,
		Outflow:      0,
	}

	return flow
}

// Adds an amount to the rate limit's flow after an incoming packet was received
// Returns an error if the new inflow will cause the rate limit to exceed its quota
func (f *Flow) AddInflow(amount uint64, quota Quota) error {
	netInflow := f.Inflow - f.Outflow + amount
	if quota.CheckExceedsQuota(PACKET_RECV, netInflow, f.ChannelValue) {
		return sdkerrors.Wrapf(ErrQuotaExceeded,
			"Inflow exceeds quota - Net Inflow: %d, Channel Value: %d, Threshold: %d%%",
			netInflow, f.ChannelValue, quota.MaxPercentRecv)
	}
	f.Inflow += amount
	return nil
}

// Adds an amount to the rate limit's flow after a packet was sent
// Returns an error if the new outflow will cause the rate limit to exceed its quota
func (f *Flow) AddOutflow(amount uint64, quota Quota) error {
	netOutflow := f.Outflow - f.Inflow + amount
	if quota.CheckExceedsQuota(PACKET_SEND, netOutflow, f.ChannelValue) {
		return sdkerrors.Wrapf(ErrQuotaExceeded,
			"Outflow exceeds quota - Net Outflow: %d, Channel Value: %d, Threshold: %d%%",
			netOutflow, f.ChannelValue, quota.MaxPercentSend)
	}
	f.Outflow += amount
	return nil
}
