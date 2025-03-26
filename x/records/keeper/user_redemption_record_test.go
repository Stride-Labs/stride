package keeper_test

import (
	"strconv"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/x/records/types"
)

// TODO [cleanup]: Migrate to new KeeperTestSuite framework
func (s *KeeperTestSuite) createNUserRedemptionRecord(n int) []types.UserRedemptionRecord {
	items := make([]types.UserRedemptionRecord, n)
	for i := range items {
		items[i].Id = strconv.Itoa(i)
		items[i].NativeTokenAmount = sdkmath.NewInt(int64(i))
		items[i].StTokenAmount = sdkmath.NewInt(int64(i))
		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestUserRedemptionRecordGet() {
	items := s.createNUserRedemptionRecord(10)
	for _, item := range items {
		got, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, item.Id)
		s.Require().True(found)
		s.Require().Equal(
			&item,
			&got,
		)
	}
}

func (s *KeeperTestSuite) TestUserRedemptionRecordRemove() {
	items := s.createNUserRedemptionRecord(10)
	for _, item := range items {
		s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, item.Id)
		_, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, item.Id)
		s.Require().False(found)
	}
}

func (s *KeeperTestSuite) TestUserRedemptionRecordGetAll() {
	items := s.createNUserRedemptionRecord(10)
	actual := s.App.RecordsKeeper.GetAllUserRedemptionRecord(s.Ctx)
	s.Require().Equal(len(items), len(actual))
}
