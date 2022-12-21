package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Initializes a new flow from the channel value
func NewFlow(channelValue sdk.Int) Flow {
	flow := Flow{
		ChannelValue: channelValue,
		Inflow:       sdk.ZeroInt(),
		Outflow:      sdk.ZeroInt(),
	}

	return flow
}

// Adds an amount to the rate limit's flow after an incoming packet was received
// Returns an error if the new inflow will cause the rate limit to exceed its quota
func (f *Flow) AddInflow(amount sdk.Int, quota Quota) error {
	netInflow := f.Inflow.Sub(f.Outflow).Add(amount)

	if quota.CheckExceedsQuota(PACKET_RECV, netInflow, f.ChannelValue) {
		return sdkerrors.Wrapf(ErrQuotaExceeded,
			"Inflow exceeds quota - Net Inflow: %d, Channel Value: %d, Threshold: %d%%",
			netInflow, f.ChannelValue, quota.MaxPercentRecv)
	}

	f.Inflow = f.Inflow.Add(amount)
	return nil
}

// Adds an amount to the rate limit's flow after a packet was sent
// Returns an error if the new outflow will cause the rate limit to exceed its quota
func (f *Flow) AddOutflow(amount sdk.Int, quota Quota) error {
	netOutflow := f.Outflow.Sub(f.Inflow).Add(amount)

	if quota.CheckExceedsQuota(PACKET_SEND, netOutflow, f.ChannelValue) {
		return sdkerrors.Wrapf(ErrQuotaExceeded,
			"Outflow exceeds quota - Net Outflow: %d, Channel Value: %d, Threshold: %d%%",
			netOutflow, f.ChannelValue, quota.MaxPercentSend)
	}

	f.Outflow = f.Outflow.Add(amount)
	return nil
}
