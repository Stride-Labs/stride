package v35_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// The handler must complete on a chain that has none of the mainnet state it acts on
// (every helper skips with a log); this is also the non-mainnet localnet case.
func (s *UpgradeTestSuite) TestUpgrade_EmptyState() {
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
}
