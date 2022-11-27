package v4_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

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
		preUpdate  func()
		update      func()
		postUpdate func()
		expPass     bool
	}{
		{
			"Test that upgrade does not panic",
			func() {
				suite.Setup()
			},
			func() {
				suite.ConfirmUpgradeSucceededs("v4", dummyUpgradeHeight)

				// make sure authz module was init
				afterCtx := suite.Ctx().WithBlockHeight(dummyUpgradeHeight)
				actGenState := suite.App.AuthzKeeper.ExportGenesis(afterCtx)
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
			tc.preUpdate()
			tc.update()
			tc.postUpdate()
		})
	}
}
