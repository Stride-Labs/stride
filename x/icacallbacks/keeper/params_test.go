package keeper_test

import (
	"github.com/Stride-Labs/stride/v31/x/icacallbacks/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	params := types.DefaultParams()

	s.App.IcacallbacksKeeper.SetParams(s.Ctx, params)

	s.Require().EqualValues(params, s.App.IcacallbacksKeeper.GetParams(s.Ctx))
}
