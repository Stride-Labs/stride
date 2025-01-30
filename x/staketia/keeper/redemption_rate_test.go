package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v25/x/staketia/types"
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
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"), // <--
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: false,
		},
		{
			name: "outside min inner",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				RedemptionRate:         sdk.MustNewDecFromStr("0.9"), // <--
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max inner",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.1"), // <--
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside min outer",
			hostZone: stakeibctypes.HostZone{
				RedemptionRate:         sdk.MustNewDecFromStr("0.8"), // <--
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.9"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			exceedsBounds: true,
		},
		{
			name: "outside max outer",
			hostZone: stakeibctypes.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.0"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.1"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.2"), // <--
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
