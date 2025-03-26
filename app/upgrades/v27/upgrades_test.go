package v27_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v26/app/upgrades/v27"
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
	upgradeHeight := int64(4)

	s.ConfirmUpgradeSucceededs(v27.UpgradeName, upgradeHeight)

	// Confirm consumer ID is set to 1
	params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(params.ConsumerId, "1")
}
