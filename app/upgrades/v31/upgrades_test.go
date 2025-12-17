package v31_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v30/app/apptesting"
	v31 "github.com/Stride-Labs/stride/v30/app/upgrades/v31"
	staketiatypes "github.com/Stride-Labs/stride/v30/x/staketia/types"
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
	// Set the staketia unbonding period to be 1
	staketiaHostZone := staketiatypes.HostZone{
		UnbondingPeriodSeconds: 1,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, staketiaHostZone)

	// Run the upgrade
	s.ConfirmUpgradeSucceeded(v31.UpgradeName)

	// Confirm it was updated to 14 days
	expectedUnbondingPeriod := uint64(1213200)
	hostZone, err := s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expectedUnbondingPeriod, hostZone.UnbondingPeriodSeconds)
}
