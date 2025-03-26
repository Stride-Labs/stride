package keeper_test

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/icacallbacks/types"
)

func (s *KeeperTestSuite) createNCallbackData(ctx sdk.Context, n int) []types.CallbackData {
	items := make([]types.CallbackData, n)
	for i := range items {
		items[i].CallbackKey = strconv.Itoa(i)

		s.App.IcacallbacksKeeper.SetCallbackData(ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestCallbackDataGet() {
	items := s.createNCallbackData(s.Ctx, 10)
	for _, item := range items {
		rst, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx,
			item.CallbackKey,
		)
		s.Require().True(found)
		s.Require().Equal(
			&item,
			&rst,
		)
	}
}

func (s *KeeperTestSuite) TestCallbackDataRemove() {
	items := s.createNCallbackData(s.Ctx, 10)
	for _, item := range items {
		s.App.IcacallbacksKeeper.RemoveCallbackData(s.Ctx,
			item.CallbackKey,
		)
		_, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx,
			item.CallbackKey,
		)
		s.Require().False(found)
	}
}

func (s *KeeperTestSuite) TestCallbackDataGetAll() {
	items := s.createNCallbackData(s.Ctx, 10)
	s.Require().ElementsMatch(
		items,
		s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx),
	)
}
