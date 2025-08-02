package v27_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v28/app/upgrades/v27"
	stakeibctypes "github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinOuterRedemptionRate sdkmath.LegacyDec
	ExpectedMaxOuterRedemptionRate sdkmath.LegacyDec
}

type UpdateRedemptionRateInnerBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinInnerRedemptionRate sdkmath.LegacyDec
	ExpectedMaxInnerRedemptionRate sdkmath.LegacyDec
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
	// Set state before upgrade
	checkLSMEnabled := s.SetupTestEnableLSM()
	checkDelegationChanges := s.SetupTestResetDelegationChangesInProgress()
	checkRedemptionRates := s.SetupTestUpdateRedemptionRateBounds()

	// Run upgrade
	s.ConfirmUpgradeSucceeded(v27.UpgradeName)

	// Confirm state after upgrade
	checkLSMEnabled()
	checkDelegationChanges()
	checkRedemptionRates()
}

func (s *UpgradeTestSuite) SetupTestEnableLSM() func() {
	// Create gaia host zone with lsm disabled
	gaiaHostZone := stakeibctypes.HostZone{
		ChainId:               v27.GaiaChainId,
		LsmLiquidStakeEnabled: false,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, gaiaHostZone)

	// Return callback to check store after upgrade
	return func() {
		gaiaHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v27.GaiaChainId)
		s.Require().True(found, "gaia found")
		s.Require().True(gaiaHostZone.LsmLiquidStakeEnabled, "lsm enabled")
	}
}

func (s *UpgradeTestSuite) SetupTestResetDelegationChangesInProgress() func() {
	// Create evmos host zone with delegation changes in progress
	evmosHostZone := stakeibctypes.HostZone{
		ChainId: v27.EvmosChainId,
		Validators: []*stakeibctypes.Validator{
			{Name: "val1", DelegationChangesInProgress: 0},
			{Name: "val2", DelegationChangesInProgress: 1},
			{Name: "val3", DelegationChangesInProgress: 2},
			{Name: "val4", DelegationChangesInProgress: 3},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, evmosHostZone)

	// Return callback to check store after upgrade
	return func() {
		expectedDelegationChanges := map[string]int64{
			"val1": 0,
			"val2": 0,
			"val3": 1,
			"val4": 2,
		}
		evmosHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v27.EvmosChainId)
		s.Require().True(found, "evmos found")

		for _, validator := range evmosHostZone.Validators {
			s.Require().Equal(expectedDelegationChanges[validator.Name],
				validator.DelegationChangesInProgress, "%s delegation change", validator.Name)
		}
	}
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v27.OsmosisChainId,
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
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
