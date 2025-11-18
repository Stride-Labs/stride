package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	stakeibctypes "github.com/Stride-Labs/stride/v30/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v30/x/staketia/types"
)

func (s *KeeperTestSuite) TestCheckRedemptionRateExceedsBounds() {
	testCases := []struct {
		name          string
		hostZone      stakeibctypes.HostZone
		exceedsBounds bool
	}{
		{
			name: "valid bounds",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.9"),
				RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.0"), // <--
				MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.2"),
			},
			exceedsBounds: false,
		},
		{
			name: "outside min inner",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.8"),
				RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("0.9"), // <--
				MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max inner",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
				RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.1"), // <--
				MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside min outer",
			hostZone: stakeibctypes.HostZone{
				RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("0.8"), // <--
				MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.9"),
				MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max outer",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.0"),
				MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.1"),
				RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.2"), // <--
			},
			exceedsBounds: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			hostZone := tc.hostZone
			hostZone.ChainId = types.CelestiaChainId
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

			err := s.App.StaketiaKeeper.CheckRedemptionRateExceedsBounds(s.Ctx)
			if tc.exceedsBounds {
				s.Require().ErrorIs(err, types.ErrRedemptionRateOutsideSafetyBounds)
			} else {
				s.Require().NoError(err, "no error expected")
			}
		})
	}
}
