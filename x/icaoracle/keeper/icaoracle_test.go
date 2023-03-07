package keeper_test

import (
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbacktypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
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
	channelId := s.CreateICAChannel(owner)
	portId, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "no error expected when formatting portId")

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
	actualCallbackData, found := s.App.ICAOracleKeeper.ICACallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
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
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, tc.Oracle.PortId, tc.Oracle.ChannelId)
	s.Require().True(found, "ica channel should have been found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, tc.Oracle.PortId, tc.Oracle.ChannelId, channel)

	// Submit the metric update which should fail
	err := s.App.ICAOracleKeeper.SubmitMetricUpdate(s.Ctx, tc.Oracle, tc.Metric)
	s.Require().ErrorContains(err, "unable to submit update oracle contract ICA: unable to submit ICA transaction")
}
