package keeper_test

import "github.com/Stride-Labs/stride/v24/x/auction/types"

func (s *KeeperTestSuite) TestParams() {
	expectedParams := types.Params{}
	err := s.App.AuctionKeeper.SetParams(s.Ctx, expectedParams)
	s.Require().NoError(err, "should not error on set params")

	actualParams, err := s.App.AuctionKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "should not error on get params")
	s.Require().Equal(expectedParams, actualParams, "params")
}
