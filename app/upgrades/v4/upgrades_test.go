package v4_test

import (
	"fmt"
	"testing"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/Stride-Labs/stride/v3/app/apptesting"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
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
				
				suite.Context = suite.Context.WithBlockHeight(dummyUpgradeHeight - 1)
				plan := upgradetypes.Plan{Name: "v4", Height: dummyUpgradeHeight}
				err := suite.App.UpgradeKeeper.ScheduleUpgrade(suite.Ctx(), plan)
				suite.Require().NoError(err)
				plan, exists := suite.App.UpgradeKeeper.GetUpgradePlan(suite.Ctx())
				suite.Require().True(exists)

				suite.Context = suite.Ctx().WithBlockHeight(dummyUpgradeHeight)
				suite.Require().NotPanics(func() {
					beginBlockRequest := abci.RequestBeginBlock{}
					suite.App.BeginBlocker(suite.Context, beginBlockRequest)
				})

				// make sure authz module was init
				actGenState := suite.App.AuthzKeeper.ExportGenesis(suite.Ctx())
				expGenState := authz.DefaultGenesisState()

				suite.Require().NotNil(actGenState)
				suite.Require().Equal(&expGenState, &actGenState)
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
