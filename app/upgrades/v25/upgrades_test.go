package v25_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v24/app/apptesting"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	// Pre upgrade setup --------------------------
	upgradeHeight := int64(4)
	checkRedemptionRatesAfterUpgrade := s.SetupTestUpdateRedemptionRateBounds()

	// Run the upgrade --------------------------
	s.ConfirmUpgradeSucceededs(v24.UpgradeName, upgradeHeight)

	// Post upgrade tests --------------------------
	checkRedemptionRatesAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v24.OsmosisChainId,
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
		},
	}

	// Create a host zone for each test case
	for _, tc := range testCases {
		hostZone := stakeibctypes.HostZone{
			ChainId:        tc.ChainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm they were all updated
		for _, tc := range testCases {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "%s - min outer", tc.ChainId)
			s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
		}
	}
}
