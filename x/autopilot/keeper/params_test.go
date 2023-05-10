package keeper_test

import (
	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	params := types.DefaultParams()
	params.StakeibcActive = false
	params.ClaimActive = true

	s.App.AutopilotKeeper.SetParams(s.Ctx, params)

	s.Require().Equal(params, s.App.AutopilotKeeper.GetParams(s.Ctx))
}
