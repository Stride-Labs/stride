package v3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
)

var (
	airdropIdentifiers = []string{"stride", "gaia", "osmosis", "juno", "stars"}
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
	s.ConfirmUpgradeSucceededs("v3", dummyUpgradeHeight)

	// make sure claim record was set
	afterCtx := s.Ctx.WithBlockHeight(dummyUpgradeHeight)
	for _, identifier := range airdropIdentifiers {
		claimRecords := s.App.ClaimKeeper.GetClaimRecords(afterCtx, identifier)
		s.Require().NotEqual(0, len(claimRecords))
	}
}
