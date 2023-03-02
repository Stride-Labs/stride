package keeper_test

import (
	"github.com/Stride-Labs/stride/v6/x/app-router/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	params := types.DefaultParams()
	params.Active = true

	s.App.RouterKeeper.SetParams(s.Ctx, params)

	s.Require().Equal(params, s.App.RouterKeeper.GetParams(s.Ctx))
}
