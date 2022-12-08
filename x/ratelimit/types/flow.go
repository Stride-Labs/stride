package types

func NewFlow(channelValue uint64) Flow {
	flow := Flow{
		ChannelValue: channelValue,
		InFlow:       0,
		OutFlow:      0,
	}

	return flow
}
