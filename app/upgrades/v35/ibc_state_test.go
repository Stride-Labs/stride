package v35_test

import (
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"

	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
)

func (s *UpgradeTestSuite) TestDeactivateICAOracles() {
	for _, oracle := range []icaoracletypes.Oracle{
		{ChainId: "injective-1", ConnectionId: "connection-1", ChannelId: "channel-1", PortId: "port-1", IcaAddress: "inj1", ContractAddress: "inj1c", Active: true},
		{ChainId: "neutron-1", ConnectionId: "connection-2", ChannelId: "channel-2", PortId: "port-2", IcaAddress: "neutron1", ContractAddress: "neutron1c", Active: true},
		{ChainId: "osmosis-1", ConnectionId: "connection-3", ChannelId: "channel-3", PortId: "port-3", IcaAddress: "osmo1", ContractAddress: "osmo1c", Active: false},
	} {
		s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)
	}

	v35.DeactivateICAOracles(s.Ctx, s.App.ICAOracleKeeper)

	oracles := s.App.ICAOracleKeeper.GetAllOracles(s.Ctx)
	s.Require().Len(oracles, 3, "no oracle is removed")
	for _, oracle := range oracles {
		s.Require().False(oracle.Active, "%s should be inactive", oracle.ChainId)
	}
}

func (s *UpgradeTestSuite) TestRemoveAllRateLimits() {
	rateLimit := func(denom, channel string) ratelimittypes.RateLimit {
		return ratelimittypes.RateLimit{
			Path:  &ratelimittypes.Path{Denom: denom, ChannelOrClientId: channel},
			Quota: &ratelimittypes.Quota{MaxPercentSend: sdkmath.NewInt(10), MaxPercentRecv: sdkmath.NewInt(10), DurationHours: 24},
			Flow:  &ratelimittypes.Flow{Inflow: sdkmath.ZeroInt(), Outflow: sdkmath.ZeroInt(), ChannelValue: sdkmath.NewInt(1000)},
		}
	}
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stuatom", "channel-0"))
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stuatom", "channel-5"))
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit("stutia", "channel-162"))
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "stuevmos")
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "stride1a", Receiver: "stride1b"})
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "stride1c", Receiver: "stride1d"})

	v35.RemoveAllRateLimits(s.Ctx, &s.App.RatelimitKeeper)

	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx), "rate limits")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx), "blacklist")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx), "whitelist")
}

func (s *UpgradeTestSuite) TestRemoveAllRateLimits_Empty() {
	s.Require().NotPanics(func() { v35.RemoveAllRateLimits(s.Ctx, &s.App.RatelimitKeeper) })
}
