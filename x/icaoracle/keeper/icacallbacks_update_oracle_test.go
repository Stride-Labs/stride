package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func (s *KeeperTestSuite) SetupTestUpdateOracleCallback() types.Metric {
	// Store an IN_PROGRESS metric
	metric := types.Metric{
		Key:               "key1",
		UpdateTime:        1,
		DestinationOracle: HostChainId,
		Status:            types.MetricStatus_IN_PROGRESS,
	}
	s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)

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
	err = s.App.ICAOracleKeeper.UpdateOracleCallback(s.Ctx, channeltypes.Packet{}, &ackResponse, callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// Confirm the pending update was removed in the case of success/failure
	expectedFound := ackStatus == icacallbacktypes.AckResponseStatus_TIMEOUT
	_, actualFound := s.App.ICAOracleKeeper.GetMetric(s.Ctx, metric.GetMetricID())
	s.Require().Equal(expectedFound, actualFound, "metric found")
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
	err := s.App.ICAOracleKeeper.UpdateOracleCallback(s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
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
	err = s.App.ICAOracleKeeper.UpdateOracleCallback(s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "metric is missing from callback")

	// Create another invalid callback args, this time with a metric struct, but no key
	callback = types.UpdateOracleCallback{
		Metric: &types.Metric{Value: "value1"},
	}
	invalidArgs, err = proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when marshalling callback data")

	// Callback should fail again
	err = s.App.ICAOracleKeeper.UpdateOracleCallback(s.Ctx, dummyPacket, &dummyAckResponse, invalidArgs)
	s.Require().ErrorContains(err, "metric is missing from callback")
}
