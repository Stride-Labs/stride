package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

func (s *KeeperTestSuite) TestCheckRedemptionRateExceedsBounds() {
	testCases := []struct {
		name          string
		hostZone      types.HostZone
		exceedsBounds bool
	}{
		{
			name: "valid bounds",
			hostZone: types.HostZone{
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
			hostZone: types.HostZone{
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
			hostZone: types.HostZone{
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
			hostZone: types.HostZone{
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
			hostZone: types.HostZone{
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
			s.App.StaketiaKeeper.SetHostZone(s.Ctx, tc.hostZone)
			err := s.App.StaketiaKeeper.CheckRedemptionRateExceedsBounds(s.Ctx)
			if tc.exceedsBounds {
				s.Require().ErrorIs(err, types.ErrRedemptionRateOutsideSafetyBounds)
			} else {
				s.Require().NoError(err, "no error expected")
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateRedemptionRateBoundsInitalized() {
	testCases := []struct {
		name     string
		hostZone types.HostZone
		valid    bool
	}{
		{
			name: "valid bounds",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: true,
		},
		{
			name: "min outer negative",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8").Neg(),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
		{
			name: "min inner negative",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9").Neg(),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
		{
			name: "max inner negative",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1").Neg(),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
		{
			name: "max outer negative",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2").Neg(),
			},
			valid: false,
		},
		{
			name: "max inner outside outer",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"),
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.3"), // <--
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
		{
			name: "min inner outside outer",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("0.7"), // <--
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"),
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
		{
			name: "min inner greater than min outer",
			hostZone: types.HostZone{
				MinRedemptionRate:      sdk.MustNewDecFromStr("0.8"),
				MinInnerRedemptionRate: sdk.MustNewDecFromStr("1.1"), // <--
				RedemptionRate:         sdk.MustNewDecFromStr("1.0"),
				MaxInnerRedemptionRate: sdk.MustNewDecFromStr("0.9"), // <--
				MaxRedemptionRate:      sdk.MustNewDecFromStr("1.2"),
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.App.StaketiaKeeper.ValidateRedemptionRateBoundsInitalized(tc.hostZone)
			if tc.valid {
				s.Require().NoError(err, "no error expected")
			} else {
				s.Require().ErrorIs(err, types.ErrInvalidRedemptionRateBounds)
			}
		})
	}
}
