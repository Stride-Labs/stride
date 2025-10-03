package v29_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	v29 "github.com/Stride-Labs/stride/v28/app/upgrades/v29"
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
	s.ConfirmUpgradeSucceeded(v29.UpgradeName)

	// Confirm state after upgrade
	consumerParams := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(consumerParams.ConsumerRedistributionFraction, "1.0")
}
