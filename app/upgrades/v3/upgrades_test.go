package v3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
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

func (suite *UpgradeTestSuite) TestUpgrade() {
	testCases := []struct {
		msg        string
		preUpdate  func()
		update     func()
		postUpdate func()
		expPass    bool
	}{
		{
			"Test that upgrade does not panic",
			func() {
				suite.Setup()
			},
			func() {
				suite.ConfirmUpgradeSucceededs("v3", dummyUpgradeHeight)

				// make sure claim record was set
				afterCtx := suite.Ctx.WithBlockHeight(dummyUpgradeHeight)
				for _, identifier := range airdropIdentifiers {
					claimRecords := suite.App.ClaimKeeper.GetClaimRecords(afterCtx, identifier)
					suite.Require().NotEqual(0, len(claimRecords))
				}
			},
			func() {
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.preUpdate()
			tc.update()
			tc.postUpdate()
		})
	}
}
