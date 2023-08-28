package keeper_test

import (
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type SubmitMetricUpdateTestCase struct {
	Oracle       types.Oracle
	Metric       types.Metric
	CallbackArgs []byte
	CallbackId   string
}

func (s *KeeperTestSuite) SetupTestSubmitMetricUpdate() SubmitMetricUpdateTestCase {
	// Create clients, connections, and an oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelId, portId := s.CreateICAChannel(owner)

	// Create oracle
	oracle := types.Oracle{
		ChainId:         HostChainId,
		ConnectionId:    ibctesting.FirstConnectionID,
		ChannelId:       channelId,
		PortId:          portId,
		IcaAddress:      "ica_address",
		ContractAddress: "contract_address",
		Active:          true,
	}
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Create metric
	metric := types.Metric{
		Key:   "key1",
		Value: "value1",
	}

	// Callback args
	callbackId := keeper.ICACallbackID_UpdateOracle
	callback := types.UpdateOracleCallback{
		OracleChainId: HostChainId,
		Metric:        &metric,
	}
	callbackBz, err := proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when serializing callback args")

	return SubmitMetricUpdateTestCase{
		Oracle:       oracle,
		Metric:       metric,
		CallbackArgs: callbackBz,
		CallbackId:   callbackId,
	}
}

func (s *KeeperTestSuite) TestSubmitMetricUpdate_Success() {
	tc := s.SetupTestSubmitMetricUpdate()

	// Call submit metric update (which should trigger an ICA)
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, tc.Oracle, tc.Metric)
	s.Require().NoError(err, "no error expected when submitting metric update")

	// Confirm callback data has been stored
	sequence := uint64(1)
	callbackKey := icacallbacktypes.PacketID(tc.Oracle.PortId, tc.Oracle.ChannelId, sequence)

	expectedCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  callbackKey,
		PortId:       tc.Oracle.PortId,
		ChannelId:    tc.Oracle.ChannelId,
		Sequence:     sequence,
		CallbackId:   tc.CallbackId,
		CallbackArgs: tc.CallbackArgs,
	}
	actualCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
	s.Require().True(found, "callback data should have been found")
	s.Require().Equal(expectedCallbackData, actualCallbackData, "callback data")
}

func (s *KeeperTestSuite) TestSubmitMetricUpdate_IcaNotRegistered() {
	tc := s.SetupTestSubmitMetricUpdate()

	// Remove ICAAddress from oracle so it appears as if the ICA was not registered
	oracle := tc.Oracle
	oracle.IcaAddress = ""

	// Submit the metric update which should fail because the ICA is not setup
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, oracle, tc.Metric)
	s.Require().ErrorContains(err, "ICAAddress is empty: oracle ICA channel has not been registered")
}

func (s *KeeperTestSuite) TestSubmitMetricUpdate_ContractNotInstantiated() {
	tc := s.SetupTestSubmitMetricUpdate()

	// Remove ContractAddress from oracle so it appears as if the contract was never instantiated
	oracle := tc.Oracle
	oracle.ContractAddress = ""

	// Submit the metric update which should fail because the contract is not instantiated
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, oracle, tc.Metric)
	s.Require().ErrorContains(err, "contract address is empty: oracle not instantiated")
}

func (s *KeeperTestSuite) TestSubmitMetricUpdate_OracleInactive() {
	tc := s.SetupTestSubmitMetricUpdate()

	// Set the oracle to inactive
	oracle := tc.Oracle
	oracle.Active = false

	// Submit the metric update which should fail because the oracle is not active
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, oracle, tc.Metric)
	s.Require().ErrorContains(err, "oracle is inactive")
}

func (s *KeeperTestSuite) TestSubmitMetricUpdate_FailedToSubmitICA() {
	tc := s.SetupTestSubmitMetricUpdate()

	// Close the channel so that the ICA fails
	s.UpdateChannelState(tc.Oracle.PortId, tc.Oracle.ChannelId, channeltypes.CLOSED)

	// Submit the metric update which should fail
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, tc.Oracle, tc.Metric)
	s.Require().ErrorContains(err, "unable to submit update oracle contract ICA: unable to send ICA tx")
}

func (s *KeeperTestSuite) TestPostAllQueuedMetrics() {
	s.SetupTestSubmitMetricUpdate()

	// Add an inactive oracle
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, types.Oracle{
		ChainId: "inactive",
		Active:  false,
	})

	// Add metrics across different states
	metrics := []types.Metric{
		// Should get sent
		{Key: "key-1", Value: "value-1", DestinationOracle: HostChainId, Status: types.MetricStatus_QUEUED},
		{Key: "key-2", Value: "value-2", DestinationOracle: HostChainId, Status: types.MetricStatus_QUEUED},
		{Key: "key-3", Value: "value-3", DestinationOracle: HostChainId, Status: types.MetricStatus_QUEUED},
		// Metric not QUEUED
		{Key: "key-4", Value: "value-4", DestinationOracle: HostChainId, Status: types.MetricStatus_IN_PROGRESS},
		// Inactive oracle - should not get sent
		{Key: "key-5", Value: "value-5", DestinationOracle: "inactive", Status: types.MetricStatus_QUEUED},
		// Oracle not found - should not get sent
		{Key: "key-6", Value: "value-6", DestinationOracle: "not-found", Status: types.MetricStatus_QUEUED},
	}
	for _, metric := range metrics {
		s.App.ICAOracleKeeper.SetMetric(s.Ctx, metric)
	}

	// Post all metrics
	s.App.ICAOracleKeeper.PostAllQueuedMetrics(s.Ctx)

	// Check 3 ICAs were submitted
	callbacks := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(callbacks, 3, "three callbacks submitted")
}
