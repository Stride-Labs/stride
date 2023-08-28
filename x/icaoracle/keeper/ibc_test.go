package keeper_test

import (
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	proto "github.com/cosmos/gogoproto/proto"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// ------------------------------------------
//				OnChanOpenAck
// ------------------------------------------

type OnChanOpenAckTestCase struct {
	ChannelId     string
	PortId        string
	ICAAddress    string
	InitialOracle types.Oracle
}

func (s *KeeperTestSuite) SetupTestOnChanOpenAck() OnChanOpenAckTestCase {
	// Create clients, connections, and an oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelId, portId := s.CreateICAChannel(owner)

	// Get ica address that was just created
	icaAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portId)
	s.Require().True(found, "ICA account should have been created")
	s.Require().NotEmpty(icaAddress, "ICA Address should not be empty")

	// Add an oracle
	oracle := types.Oracle{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
	}
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Confirm the oracle was stored
	_, found = s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be in the store during setup")

	return OnChanOpenAckTestCase{
		ChannelId:     channelId,
		PortId:        portId,
		ICAAddress:    icaAddress,
		InitialOracle: oracle,
	}
}

func (s *KeeperTestSuite) TestOnChanOpenAck_Success() {
	tc := s.SetupTestOnChanOpenAck()

	// Call callback
	err := s.App.ICAOracleKeeper.OnChanOpenAck(s.Ctx, tc.PortId, tc.ChannelId)
	s.Require().NoError(err, "no error expected when calling OnChanOpenAck")

	// Confirm oracle was updated
	expectedOracle := tc.InitialOracle
	expectedOracle.ChannelId = tc.ChannelId
	expectedOracle.PortId = tc.PortId
	expectedOracle.IcaAddress = tc.ICAAddress

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should have been found")
	s.Require().Equal(expectedOracle, actualOracle, "oracle should have updated")
}

func (s *KeeperTestSuite) TestOnChanOpenAck_ConnectionNotFound() {
	tc := s.SetupTestOnChanOpenAck()

	// Pass a different channel-id - the connection should not be found and the callback should error
	err := s.App.ICAOracleKeeper.OnChanOpenAck(s.Ctx, tc.PortId, "fake_channel")
	s.Require().ErrorContains(err, "unable to get connection from channel (fake_channel) and port")
}

func (s *KeeperTestSuite) TestOnChanOpenAck_NoOracle() {
	tc := s.SetupTestOnChanOpenAck()

	// Update the oracle to have a different  connection so it cannoth be found
	oracle := tc.InitialOracle
	oracle.ConnectionId = "different_connection_id"
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// The callback should not fail (as it can be called by non-oracle callbacks)
	// But the oracle should not be updated
	err := s.App.ICAOracleKeeper.OnChanOpenAck(s.Ctx, tc.PortId, tc.ChannelId)
	s.Require().NoError(err, "no error expected when calling OnChanOpenAck")

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should have been found")
	s.Require().Equal(oracle, actualOracle, "oracle should not have updated")
}

func (s *KeeperTestSuite) TestOnChanOpenAck_NotOracleChannel() {
	tc := s.SetupTestOnChanOpenAck()

	// Create non-oracle ICA channel and use that for the callback
	owner := types.FormatICAAccountOwner(HostChainId, "NOT_ORACLE")
	differentChannelId, differentPortId := s.CreateICAChannel(owner)

	// The callback should succeed but the oracle should not be updated
	err := s.App.ICAOracleKeeper.OnChanOpenAck(s.Ctx, differentPortId, differentChannelId)
	s.Require().NoError(err, "no error expected when calling OnChanOpenAck")

	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should have been found")
	s.Require().Equal(tc.InitialOracle, actualOracle, "oracle should not have updated")
}

func (s *KeeperTestSuite) TestOnChanOpenAck_NoICAAddress() {
	tc := s.SetupTestOnChanOpenAck()

	// Update the oracle's channel/port to map to a different connection
	differentConnectionId := "connection-2"
	connection := connectiontypes.ConnectionEnd{}
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, differentConnectionId, connection)

	channel := channeltypes.Channel{
		ConnectionHops: []string{differentConnectionId},
	}
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, tc.PortId, tc.ChannelId, channel)

	// Update the oracle struct to use the different connection as well
	oracle := tc.InitialOracle
	oracle.ConnectionId = differentConnectionId
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	err := s.App.ICAOracleKeeper.OnChanOpenAck(s.Ctx, tc.PortId, tc.ChannelId)
	s.Require().ErrorContains(err, "unable to get ica address from connection")
}

// ------------------------------------------
//				SubmitICATx
// ------------------------------------------

func (s *KeeperTestSuite) SetupTestSubmitICATx() (tx types.ICATx, callbackBz []byte) {
	// Create clients, connections, and an oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelId, portId := s.CreateICAChannel(owner)

	// Callback args (we can use any callback type here)
	callback := types.InstantiateOracleCallback{OracleChainId: HostChainId}
	callbackBz, err := proto.Marshal(&callback)
	s.Require().NoError(err, "no error expected when serializing callback args")

	// Return a valid ICATx
	return types.ICATx{
		ConnectionId:    ibctesting.FirstConnectionID,
		ChannelId:       channelId,
		PortId:          portId,
		Owner:           owner,
		Messages:        []proto.Message{&banktypes.MsgSend{}},
		RelativeTimeout: time.Second,
		CallbackId:      "callback_id",
		CallbackArgs:    &callback,
	}, callbackBz
}

func (s *KeeperTestSuite) TestSubmitICATx_Success() {
	icaTx, callbackBz := s.SetupTestSubmitICATx()

	// Submit ICA
	err := s.App.ICAOracleKeeper.SubmitICATx(s.Ctx, icaTx)
	s.Require().NoError(err, "no error expected when submitting ICA")

	// Confirm callback data was stored
	sequence := uint64(1)
	callbackKey := icacallbacktypes.PacketID(icaTx.PortId, icaTx.ChannelId, sequence)

	expectedCallbackData := icacallbacktypes.CallbackData{
		CallbackKey:  callbackKey,
		PortId:       icaTx.PortId,
		ChannelId:    icaTx.ChannelId,
		Sequence:     sequence,
		CallbackId:   icaTx.CallbackId,
		CallbackArgs: callbackBz,
	}
	actualCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
	s.Require().True(found, "callback data should have been found")
	s.Require().Equal(expectedCallbackData, actualCallbackData, "callback data")
}

func (s *KeeperTestSuite) TestSubmitICATx_InvalidICATx() {
	icaTx, _ := s.SetupTestSubmitICATx()

	// Submit ICA without a connection-id - should fail
	icaTx.ConnectionId = ""
	err := s.App.ICAOracleKeeper.SubmitICATx(s.Ctx, icaTx)
	s.Require().ErrorContains(err, "connection-id is empty: invalid ICA request")
}

func (s *KeeperTestSuite) TestSubmitICATx_InvalidMessage() {
	icaTx, _ := s.SetupTestSubmitICATx()

	// Submit ICA without a nil message - should fail
	icaTx.Messages = []proto.Message{nil}
	err := s.App.ICAOracleKeeper.SubmitICATx(s.Ctx, icaTx)
	s.Require().ErrorContains(err, "unable to serialize cosmos transaction")
}

func (s *KeeperTestSuite) TestSubmitICATx_SendFailure() {
	icaTx, _ := s.SetupTestSubmitICATx()

	// Close the channel so that the ICA fails
	s.UpdateChannelState(icaTx.PortId, icaTx.ChannelId, channeltypes.CLOSED)

	// Submit the ICA which should error
	err := s.App.ICAOracleKeeper.SubmitICATx(s.Ctx, icaTx)
	s.Require().ErrorContains(err, "unable to send ICA tx")
}
