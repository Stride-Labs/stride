package keeper_test

import (
	"github.com/Stride-Labs/stride/v31/x/icacallbacks/types"
)

func (s *KeeperTestSuite) TestParamsQuery() {
	params := types.DefaultParams()
	s.App.IcacallbacksKeeper.SetParams(s.Ctx, params)

	response, err := s.App.IcacallbacksKeeper.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
