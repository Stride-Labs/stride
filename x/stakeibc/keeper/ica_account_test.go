package keeper_test

import (
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (s *KeeperTestSuite) createTestICAAccount() types.ICAAccount {
	item := types.ICAAccount{}
	s.App.StakeibcKeeper.SetICAAccount(s.Ctx(), item)
	return item
}

func (s *KeeperTestSuite) TestICAAccountGet() {
	item := s.createTestICAAccount()
	rst, found := s.App.StakeibcKeeper.GetICAAccount(s.Ctx())
	s.Require().True(found)
	s.Require().Equal(&item, &rst)
}

func (s *KeeperTestSuite) TestICAAccountRemove() {
	s.createTestICAAccount()
	s.App.StakeibcKeeper.RemoveICAAccount(s.Ctx())
	_, found := s.App.StakeibcKeeper.GetICAAccount(s.Ctx())
	s.Require().False(found)
}
