package v15_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	v15 "github.com/Stride-Labs/stride/v14/app/upgrades/v15"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

type UpdateRedemptionRateBounds struct {
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMinInnerRedemptionRate sdk.Dec
	ExpectedMaxInnerRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("0.97"), // 1 - 3% = 0.97
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.05"), // 1 + 5% = 1.05
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.067"), // 1.1 - 3% = 1.067
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.155"), // 1.1 + 5% = 1.155
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.2125"), // 1.25 - 3% = 1.2125
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.3125"), // 1.25 + 5% = 1.3125
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.3750"), // 1.25 + 10% = 1.375
		},
	}

	// Create a host zone for each test case
	for i, tc := range testCases {
		chainId := fmt.Sprintf("chain-%d", i)

		hostZone := stakeibctypes.HostZone{
			ChainId:        chainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Create an evmos host zone
	initialEvmosMaxOuter := sdk.MustNewDecFromStr("1.50")
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:           v15.EvmosChainId,
		MaxRedemptionRate: initialEvmosMaxOuter,
	})

	// Run the upgrade to set the bounds
	s.ConfirmUpgradeSucceededs("v15", dummyUpgradeHeight)

	// Confirm the correct bounds were set
	for i, tc := range testCases {
		chainId := fmt.Sprintf("chain-%d", i)

		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(found)

		s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "min outer")
		s.Require().Equal(tc.ExpectedMinInnerRedemptionRate, hostZone.MinInnerRedemptionRate, "min inner")
		s.Require().Equal(tc.ExpectedMaxInnerRedemptionRate, hostZone.MinInnerRedemptionRate, "max inner")
		s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "max outer")
	}

	// Confirm evmos' custom bounds were set
	evmosHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v15.EvmosChainId)
	s.Require().True(found)

	s.Require().Equal(v15.EvmosOuterMinRedemptionRate, evmosHostZone.MinRedemptionRate, "min outer")
	s.Require().Equal(v15.EvmosInnerMinRedemptionRate, evmosHostZone.MinInnerRedemptionRate, "min inner")
	s.Require().Equal(initialEvmosMaxOuter, evmosHostZone.MinInnerRedemptionRate, "max inner")
	s.Require().Equal(initialEvmosMaxOuter, evmosHostZone.MaxRedemptionRate, "max outer")
}
