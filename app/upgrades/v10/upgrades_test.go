package v10_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	v10 "github.com/Stride-Labs/stride/v9/app/upgrades/v10"
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

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	s.ConfirmUpgradeSucceededs("v10", dummyUpgradeHeight)

	// Check mint parameters after upgrade
	zeroPad := "00000000000000"
	proportions := s.App.MintKeeper.GetParams(s.Ctx).DistributionProportions

	s.Require().Equal(v10.StakingProportion+zeroPad,
		proportions.Staking.String(), "staking")

	s.Require().Equal(v10.CommunityPoolGrowthProportion+zeroPad,
		proportions.CommunityPoolGrowth.String(), "community pool growth")

	s.Require().Equal(v10.StrategicReserveProportion+zeroPad,
		proportions.StrategicReserve.String(), "strategic reserve")

	s.Require().Equal(v10.CommunityPoolSecurityBudgetProportion+zeroPad,
		proportions.CommunityPoolSecurityBudget.String(), "community pool security")

	// Check initial deposit ratio
	govParams := s.App.GovKeeper.GetParams(s.Ctx)
	s.Require().Equal(v10.MinInitialDepositRatio, govParams.MinInitialDepositRatio, "min initial deposit ratio")
}
