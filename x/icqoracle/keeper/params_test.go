package keeper_test

import "github.com/Stride-Labs/stride/v29/x/icqoracle/types"

func (s *KeeperTestSuite) TestParams() {
	expectedParams := types.Params{
		OsmosisChainId:            "osmosis-1",
		OsmosisConnectionId:       "connection-2",
		UpdateIntervalSec:         5 * 60,  // 5 min
		PriceExpirationTimeoutSec: 15 * 60, // 15 min
	}
	s.App.ICQOracleKeeper.SetParams(s.Ctx, expectedParams)

	actualParams := s.App.ICQOracleKeeper.GetParams(s.Ctx)
	s.Require().Equal(expectedParams, actualParams, "params")
}
