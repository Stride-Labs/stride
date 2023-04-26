package keeper_test

import (
	"context"

	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
)

func (s *KeeperTestSuite) TestParamsQuery() {
	// Test with stakeibc enabled and claim disabled
	s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{
		StakeibcActive: true,
		ClaimActive:    false,
	})
	queryResponse, err := s.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().True(queryResponse.Params.StakeibcActive)
	s.Require().False(queryResponse.Params.ClaimActive)

	// Test with claim enabled and stakeibc disabled
	s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{
		StakeibcActive: false,
		ClaimActive:    true,
	})
	queryResponse, err = s.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().False(queryResponse.Params.StakeibcActive)
	s.Require().True(queryResponse.Params.ClaimActive)
}
