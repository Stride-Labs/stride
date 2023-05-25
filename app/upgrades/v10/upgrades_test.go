package v10_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
)

var (
	dummyUpgradeHeight = int64(5)
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) TestUpgrade() {
	suite.Setup()
	suite.ConfirmUpgradeSucceededs("v10", dummyUpgradeHeight)
}
