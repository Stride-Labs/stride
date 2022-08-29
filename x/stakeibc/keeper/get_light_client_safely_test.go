package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
)

type GetLightClientSafelyTestCase struct {
	connectionId string
}

func (s *KeeperTestSuite) SetupGetLightClientSafelyTokens() GetLightClientSafelyTestCase {
	connectionId := "connection-0"

	return GetLightClientSafelyTestCase{
		connectionId: connectionId,
	}
}
func (s *KeeperTestSuite) TestGetLightClientSafelyTokensSuccessful() {
	tc := s.SetupGetLightClientSafelyTokens()

	actualLightClientTime, err := s.App.StakeibcKeeper.GetLightClientTimeSafely(s.Ctx, tc.connectionId)
	s.Require().NoError(err, "light client time could be fetched")
	s.Require().Equal(0, int(actualLightClientTime), "light client time could be fetched")
	s.Require().False(true, "light client time could be fetched")
}

// func (s *KeeperTestSuite) TestGetLightClientSafelyTokensNoUserRedemptionRecord() {
// 	tc := s.SetupGetLightClientSafelyTokens()
// 	// Remove the user redemption record
// 	s.App.RecordsKeeper.RemoveUserRedemptionRecord(s.Ctx, tc.initialState.redemptionRecordId)

// 	_, err := s.msgServer.GetLightClientSafelyTokens(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
// 	expectedErr := "unable to find claimable redemption record: "
// 	expectedErr += "could not get user redemption record: GAIA.1.stride_SENDER: user redemption record error"
// 	s.Require().EqualError(err, expectedErr)
// }
