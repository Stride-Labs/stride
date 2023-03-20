package keeper_test

import (
	"context"

	"github.com/Stride-Labs/stride/v7/x/autopilot/types"
)

func (s *KeeperTestSuite) TestParamsQuery() {
	// Test with app-route param active
	s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{Active: true})
	queryResponse, err := s.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().True(queryResponse.Params.Active)

	// Test with app-route param in-active
	s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{Active: false})
	queryResponse, err = s.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().False(queryResponse.Params.Active)
}
