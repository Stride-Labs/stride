package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestIsRedemptionRateWithinSafetyBounds() {
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.DefaultMinRedemptionRateThreshold = 75
	params.DefaultMaxRedemptionRateThreshold = 150
	hostZones := make(map[string]types.HostZone)
	hostZones["gaia-1"] = types.HostZone{
		ChainId:           "gaia-1",
		MinRedemptionRate: sdkmath.LegacyNewDecWithPrec(15, 1), // 1.5
		MaxRedemptionRate: sdkmath.LegacyNewDecWithPrec(25, 1), // 2.5
	}
	hostZones["osmosis-1"] = types.HostZone{
		ChainId:           "osmosis-1",
		MinRedemptionRate: sdkmath.LegacyNewDecWithPrec(3, 1),  // 0.3
		MaxRedemptionRate: sdkmath.LegacyNewDecWithPrec(20, 1), // 2
	}
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	for _, tc := range []struct {
		chainId        string
		redemptionRate sdkmath.LegacyDec
		expSafe        bool
	}{
		{
			chainId:        "osmosis-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(3, 1), // 0.3
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "osmosis-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "gaia-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(25, 1), // 2.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(1, 1), // 0.1
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(3, 1), // 0.3
			expSafe:        false,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(15, 1), // 1.5
			expSafe:        true,
		},
		{
			chainId:        "stars-1",
			redemptionRate: sdkmath.LegacyNewDecWithPrec(25, 1), // 2.5
			expSafe:        false,
		},
	} {
		hostZone, ok := hostZones[tc.chainId]
		if !ok {
			hostZone = types.HostZone{
				ChainId: tc.chainId,
			}
		}
		hostZone.RedemptionRate = tc.redemptionRate
		rrSafe, err := s.App.StakeibcKeeper.IsRedemptionRateWithinSafetyBounds(s.Ctx, hostZone)
		if tc.expSafe {
			s.Require().NoError(err)
			s.Require().True(rrSafe)
		} else {
			s.Require().Error(err)
			s.Require().False(rrSafe)
		}
	}
}
