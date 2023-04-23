package v4_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
)

const dummyUpgradeHeight = 5

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	suite.Setup()
	suite.ConfirmUpgradeSucceededs("v4", dummyUpgradeHeight)
}
