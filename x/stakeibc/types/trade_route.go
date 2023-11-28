package types

func (t TradeRoute) GetKey() []byte {
	return TradeRouteKeyFromDenoms(t.RewardDenomOnRewardZone, t.HostDenomOnHostZone)
}
