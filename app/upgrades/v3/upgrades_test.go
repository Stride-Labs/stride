package v3_test

import (
	"fmt"
	"testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/Stride-Labs/stride/v3/app/apptesting"
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
		msg         string
		pre_update  func()
		update      func()
		post_update func()
		expPass     bool
	}{
		{
			"Test that upgrade does not panic",
			func() {
				// Create pool 1
				suite.Setup()
			},
			func() {
				// run upgrade
				// TODO: Refactor this all into a helper fn
				
				beforeCtx := suite.Ctx().WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v3", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(beforeCtx, plan)
				suite.Require().NoError(err)
				plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(beforeCtx)
				suite.Require().True(exists)

				afterCtx := suite.Ctx().WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(afterCtx, beginBlockRequest)
				})

				// make sure claim record was set
				for _, identifier := range(airdropIdentifiers) {
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
			tc.pre_update()
			tc.update()
			tc.post_update()
		})
	}
}
