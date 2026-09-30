package v35_test

import (
	sdkmath "cosmossdk.io/math"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	stakedymtypes "github.com/Stride-Labs/stride/v34/x/stakedym/types"
)

// seedHaltedStakedym stores the stakedym zone as it is on mainnet: halted, with its rate
// (1.100538) above both max bounds (1.1)
func (s *UpgradeTestSuite) seedHaltedStakedym() {
	s.App.StakedymKeeper.SetHostZone(s.Ctx, stakedymtypes.HostZone{
		ChainId:                "dymension_1100-1",
		NativeTokenDenom:       "adym",
		RedemptionRate:         sdkmath.LegacyMustNewDecFromStr("1.100538"),
		MinRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("0.9"),
		MinInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"),
		MaxInnerRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1"),
		MaxRedemptionRate:      sdkmath.LegacyMustNewDecFromStr("1.1"),
		Halted:                 true,
	})
}

func (s *UpgradeTestSuite) TestUnhaltStakedym() {
	s.seedHaltedStakedym()

	v35.UnhaltStakedym(s.Ctx, s.App.StakedymKeeper)

	hostZone, err := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().False(hostZone.Halted, "unhalted")
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("2.201076"), hostZone.MaxRedemptionRate, "outer max widened")
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("2.201076"), hostZone.MaxInnerRedemptionRate, "inner max widened")
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("0.9"), hostZone.MinRedemptionRate, "min untouched")
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("0.95"), hostZone.MinInnerRedemptionRate, "inner min untouched")
	s.Require().NoError(s.App.StakedymKeeper.CheckRedemptionRateExceedsBounds(s.Ctx), "the BeginBlocker check passes")
}

func (s *UpgradeTestSuite) TestUnhaltStakedym_KeepsWiderBounds() {
	s.seedHaltedStakedym()
	hostZone, _ := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	hostZone.MaxRedemptionRate = sdkmath.LegacyMustNewDecFromStr("5")
	hostZone.MaxInnerRedemptionRate = sdkmath.LegacyMustNewDecFromStr("3")
	s.App.StakedymKeeper.SetHostZone(s.Ctx, hostZone)

	v35.UnhaltStakedym(s.Ctx, s.App.StakedymKeeper)

	after, _ := s.App.StakedymKeeper.GetHostZone(s.Ctx)
	s.Require().False(after.Halted)
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("5"), after.MaxRedemptionRate, "already wide enough")
	s.Require().Equal(sdkmath.LegacyMustNewDecFromStr("3"), after.MaxInnerRedemptionRate, "already wide enough")
}

func (s *UpgradeTestSuite) TestUnhaltStakedym_NoZone() {
	s.Require().NotPanics(func() { v35.UnhaltStakedym(s.Ctx, s.App.StakedymKeeper) })
}
