package keeper_test

import (
	"github.com/Stride-Labs/stride/v28/x/records/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	params := types.DefaultParams()

	s.App.RecordsKeeper.SetParams(s.Ctx, params)

	s.Require().EqualValues(params, s.App.RecordsKeeper.GetParams(s.Ctx))
}
