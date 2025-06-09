package keeper_test

import "github.com/Stride-Labs/stride/v27/x/airdrop/types"

func (s *KeeperTestSuite) TestParams() {
	expectedParams := types.Params{PeriodLengthSeconds: 24 * 60 * 60}
	s.App.AirdropKeeper.SetParams(s.Ctx, expectedParams)
	actualParams := s.App.AirdropKeeper.GetParams(s.Ctx)
	s.Require().Equal(expectedParams, actualParams, "params")
}
