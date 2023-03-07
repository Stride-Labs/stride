package keeper_test

import (
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func (s *KeeperTestSuite) SetupTestUpdateOracleCallback() types.Metric {
	// Store pending metric update
	metric := types.Metric{
		Key:        "key1",
		UpdateTime: 1,
	}
	s.App.ICAOracleKeeper.SetMetricUpdateInProgress(s.Ctx, types.PendingMetricUpdate{
		Metric:        &metric,
		OracleChainId: HostChainId,
	})

	// Confirm update is stored
	_, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metric.Key, HostChainId, metric.UpdateTime)
	s.Require().True(found, "pending metric update should be in the store during setup")

	return metric
}

func (s *KeeperTestSuite) CallCallbackAndCheckState(ackStatus icacallbacktypes.AckResponseStatus) {
	metric := s.SetupTestUpdateOracleCallback()

	// Serialize callback
	callback := types.UpdateOracleCallback{
		Metric: &metric,
	}
	callbackBz, err := proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Call update oracle callback
	ackResponse := icacallbacktypes.AcknowledgementResponse{
		Status: ackStatus,
	}
	keeper.UpdateOracleCallback(s.App.ICAOracleKeeper, s.Ctx, channeltypes.Packet{}, &ackResponse, callbackBz)

	// Confirm the pending update was removed
	_, found := s.App.ICAOracleKeeper.GetPendingMetricUpdate(s.Ctx, metric.Key, HostChainId, metric.UpdateTime)
	s.Require().True(found, "pending metric update should have been removed")
}

func (s *KeeperTestSuite) TestUpdateOracleCallback_AckSuccess() {
	s.CallCallbackAndCheckState(icacallbacktypes.AckResponseStatus_SUCCESS)
}

func (s *KeeperTestSuite) TestUpdateOracleCallback_AckTimeout() {
	s.CallCallbackAndCheckState(icacallbacktypes.AckResponseStatus_TIMEOUT)
}

func (s *KeeperTestSuite) TestUpdateOracleCallback_AckFailure() {
	s.CallCallbackAndCheckState(icacallbacktypes.AckResponseStatus_FAILURE)
}

func (s *KeeperTestSuite) TestUpdateOracleCallback_UnmarshalFailure() {
	dummyPacket := channeltypes.Packet{}
	dummyAckResponse := icacallbacktypes.AcknowledgementResponse{}
	invalidArgs := []byte{1, 2, 3}
	err := keeper.UpdateOracleCallback(s.App.ICAOracleKeeper, s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "unable to unmarshal update oracle callback")
}

func (s *KeeperTestSuite) TestUpdateOracleCallback_InvalidCallbackData() {
	dummyPacket := channeltypes.Packet{}
	dummyAckResponse := icacallbacktypes.AcknowledgementResponse{}

	// Create invalid callback args with no metric field
	callback := types.UpdateOracleCallback{}
	invalidArgs, err := proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Callback should fail
	err = keeper.UpdateOracleCallback(s.App.ICAOracleKeeper, s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "metric is missing from callback")

	// Create another invalid callback args, this time with a metric struct, but no key
	callback = types.UpdateOracleCallback{
		Metric: &types.Metric{Value: "value1"},
	}
	invalidArgs, err = proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Callback should fail again
	err = keeper.UpdateOracleCallback(s.App.ICAOracleKeeper, s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "metric is missing from callback")
}
