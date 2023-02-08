package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
)

type GetLightClientSafelyTestCase struct {
	connectionId              string
	expectedLightClientTime   int64
	expectedLightClientHeight int64
}

func (s *KeeperTestSuite) SetupGetLightClientSafely() GetLightClientSafelyTestCase {
	connectionId := "connection-0"
	s.CreateTransferChannel("GAIA")

	// note this time is Jan 2020, set in the ibc test setup
	expectedLightClientTime := int64(1577923340000000000)
	// note this is the block height post-setup in the ibc test setup (creating connections, channels etc advances the block)
	//        this may change as we ament the setup, please update accordingly!
	expectedLightClientHeight := int64(16)

	return GetLightClientSafelyTestCase{
		connectionId:              connectionId,
		expectedLightClientTime:   expectedLightClientTime,
		expectedLightClientHeight: expectedLightClientHeight,
	}
}

func (s *KeeperTestSuite) TestGetLightClientTimeSafely_Successful() {
	tc := s.SetupGetLightClientSafely()

	actualLightClientTime, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientTime), 0, "light client time g.t. 0")
	s.Require().Equal(tc.expectedLightClientTime, int64(actualLightClientTime), "light client time matches expected time")

	// update LC to new block on host chain
	//   NOTE this advances the time!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")
	timeDelta := 10000000000

	actualLightClientTimeNewTime, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "new light client time could be fetched")

	s.Require().Equal(int64(actualLightClientTimeNewTime), int64(actualLightClientTime+uint64(timeDelta)), "light client time increments by expected amount")
}

func (s *KeeperTestSuite) TestGetLightClientSafely_InvalidConnection() {
	tc := s.SetupGetLightClientSafely()
	tc.connectionId = "connection-invalid"

	_, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "invalid connection id", "get lc time: error complains about invalid connection id")

	_, err = s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().ErrorContains(err, "invalid connection id", "get lc height: error complains about invalid connection id")
}

func (s *KeeperTestSuite) TestGetLightClientHeightSafely_Successful() {
	tc := s.SetupGetLightClientSafely()

	actualLightClientHeight, err := s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Greater(int(actualLightClientHeight), 0, "light client height g.t. 0")
	s.Require().Equal(int64(actualLightClientHeight), tc.expectedLightClientHeight, "light client height matches expected height")

	// update LC to new block on host chain
	//   NOTE this advances the block height!
	err = s.TransferPath.EndpointA.UpdateClient()
	s.Require().NoError(err, "update client")

	actualLightClientHeightNextBlock, err := s.App.StakeibcKeeper.GetLightClientHeightSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")

	s.Require().Equal(int64(actualLightClientHeightNextBlock), int64(actualLightClientHeight+1), "light client height matches expected height")
}
