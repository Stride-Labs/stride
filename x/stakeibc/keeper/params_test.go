package keeper_test

import (
	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	params := types.DefaultParams()

	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	s.Require().EqualValues(params, s.App.StakeibcKeeper.GetParams(s.Ctx))
}
