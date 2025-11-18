package keeper_test

import (
	"github.com/Stride-Labs/stride/v30/x/records/types"
)

func (s *KeeperTestSuite) TestParamsQuery() {
	params := types.DefaultParams()
	s.App.RecordsKeeper.SetParams(s.Ctx, params)

	response, err := s.App.RecordsKeeper.Params(s.Ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsResponse{Params: params}, response)
}
