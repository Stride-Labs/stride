package types

import (
	"errors"
	fmt "fmt"
)

func NewFlow(channelValue uint64) Flow {
	flow := Flow{
		ChannelValue: channelValue,
		Inflow:       0,
		Outflow:      0,
	}

	return flow
}

func (f *Flow) AddInflow(amount uint64, quota Quota) error {
	netInflow := f.Inflow - f.Outflow + amount
	if quota.CheckExceedsQuota(PACKET_RECV, netInflow, f.ChannelValue) {
		return errors.New(fmt.Sprintf("Inflow exceeds quota - Net Inflow: %d, Channel Value: %d, Threshold: %d%%",
			netInflow, f.ChannelValue, quota.MaxPercentRecv))
	}
	f.Inflow += amount
	return nil
}

func (f *Flow) AddOutflow(amount uint64, quota Quota) error {
	netOutflow := f.Outflow - f.Inflow + amount
	if quota.CheckExceedsQuota(PACKET_SEND, netOutflow, f.ChannelValue) {
		return errors.New(fmt.Sprintf("Outflow exceeds quota - Net Outflow: %d, Channel Value: %d, Threshold: %d%%",
			netOutflow, f.ChannelValue, quota.MaxPercentSend))
	}
	f.Outflow += amount
	return nil
}
