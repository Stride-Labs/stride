package keeper_test

import "github.com/Stride-Labs/stride/v25/x/icqoracle/types"

func (s *KeeperTestSuite) TestParams() {
	expectedParams := types.Params{
		OsmosisChainId:            "osmosis-1",
		OsmosisConnectionId:       "connection-2",
		UpdateIntervalSec:         5 * 60,  // 5 min
		PriceExpirationTimeoutSec: 15 * 60, // 15 min
	}
	err := s.App.ICQOracleKeeper.SetParams(s.Ctx, expectedParams)
	s.Require().NoError(err, "should not error on set params")

	actualParams, err := s.App.ICQOracleKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "should not error on get params")
	s.Require().Equal(expectedParams, actualParams, "params")
}
