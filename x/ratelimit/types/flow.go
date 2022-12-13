package types

func NewFlow(channelValue uint64) Flow {
	flow := Flow{
		ChannelValue: channelValue,
		Inflow:       0,
		Outflow:      0,
	}

	return flow
}
