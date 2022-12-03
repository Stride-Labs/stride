package keeper_test

import (
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (suite *KeeperTestSuite) createTestICAAccount() types.ICAAccount {
	item := types.ICAAccount{}
	suite.App.StakeibcKeeper.SetICAAccount(suite.Ctx, item)
	return item
}

func (suite *KeeperTestSuite) TestICAAccountGet() {
	item := suite.createTestICAAccount()
	rst, found := suite.App.StakeibcKeeper.GetICAAccount(suite.Ctx)
	suite.Require().True(found)
	suite.Require().Equal(&item, &rst)
}

func (suite *KeeperTestSuite) TestICAAccountRemove() {
	suite.createTestICAAccount()
	suite.App.StakeibcKeeper.RemoveICAAccount(suite.Ctx)
	_, found := suite.App.StakeibcKeeper.GetICAAccount(suite.Ctx)
	suite.Require().False(found)
}
