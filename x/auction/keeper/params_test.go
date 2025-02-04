package keeper_test

import "github.com/Stride-Labs/stride/v25/x/auction/types"

func (s *KeeperTestSuite) TestParams() {
	expectedParams := types.Params{}
	s.App.AuctionKeeper.SetParams(s.Ctx, expectedParams)

	actualParams := s.App.AuctionKeeper.GetParams(s.Ctx)
	s.Require().Equal(expectedParams, actualParams, "params")
}
