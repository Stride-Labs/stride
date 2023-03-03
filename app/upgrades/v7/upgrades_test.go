package v7_test

import (
	"testing"

	//nolint:staticcheck

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v6/app/apptesting"
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

func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	// add tests
}
