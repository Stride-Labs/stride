package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type RestoreOracleICATestCase struct {
	ValidMsg types.MsgRestoreOracleICA
	Oracle   types.Oracle
}

func (s *KeeperTestSuite) SetupTestRestoreOracleICA() RestoreOracleICATestCase {
	// Create oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelId, portId := s.CreateICAChannel(owner)

	// Create oracle
	oracle := types.Oracle{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		ChannelId:    channelId,
		PortId:       portId,
		IcaAddress:   "ica_address",
	}
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Confirm the oracle was stored
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be in the store during setup")

	// Close the channel (to test the restore functionality)
	s.UpdateChannelState(portId, channelId, channeltypes.CLOSED)

	return RestoreOracleICATestCase{
		ValidMsg: types.MsgRestoreOracleICA{OracleChainId: HostChainId},
		Oracle:   oracle,
	}
}

func (s *KeeperTestSuite) TestRestoreOracleICA_Successful() {
	tc := s.SetupTestRestoreOracleICA()

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + oracle)")

	// Submit the restore message
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().NoError(err, "no error expected when restoring an oracle ICA")

	// Confirm the new channel was created
	channels = s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 3, "there should be 3 channels after restoring")

	// Confirm the new channel is in state INIT
	newChannelActive := false
	for _, channel := range channels {
		// The new channel should have the same port, a new channel ID and be in state INIT
		if channel.PortId == tc.Oracle.PortId && channel.ChannelId != tc.Oracle.ChannelId && channel.State == channeltypes.INIT {
			newChannelActive = true
		}
	}
	s.Require().True(newChannelActive, "a new channel should have been created")
}

func (s *KeeperTestSuite) TestRestoreOracleICA_OracleDoesNotExist() {
	tc := s.SetupTestRestoreOracleICA()

	// Submit the oracle with an invalid host zone, it should fail
	invalidMsg := tc.ValidMsg
	invalidMsg.OracleChainId = "fake_chain"
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "oracle not found")
}

func (s *KeeperTestSuite) TestRestoreOracleICA_IcaNotRegistered() {
	tc := s.SetupTestRestoreOracleICA()

	// Update the oracle to appear as if the ICA was never registered in the first place
	oracle := tc.Oracle
	oracle.IcaAddress = ""
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Submit the restore message - it should fail
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, fmt.Sprintf("the oracle (%s) has never had an registered ICA", HostChainId))
}

func (s *KeeperTestSuite) TestRestoreOracleICA_ConnectionDoesNotExist() {
	tc := s.SetupTestRestoreOracleICA()

	// Update the oracle to to have a non-existent connection-id
	oracle := tc.Oracle
	oracle.ConnectionId = "fake_connection"
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Submit the rsetore message - it should fail
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "connection (fake_connection) not found")
}

func (s *KeeperTestSuite) TestRestoreOracleICA_Failure_IcaDoesNotExist() {
	tc := s.SetupTestRestoreOracleICA()

	// Add a new connection-id that is not tied to an ICA
	differentConnectionId := "connection-2"
	connection := connectiontypes.ConnectionEnd{}
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, differentConnectionId, connection)

	// Update the oracle to have that connectionId
	oracle := tc.Oracle
	oracle.ConnectionId = differentConnectionId
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Submit the restore message - it should fail
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "cannot find ICA account for connection")
}

func (s *KeeperTestSuite) TestRestoreOracleICA_Failure_ChannelOpen() {
	tc := s.SetupTestRestoreOracleICA()

	// Open the channel back up
	s.UpdateChannelState(tc.Oracle.PortId, tc.Oracle.ChannelId, channeltypes.OPEN)

	// Since the channel already OPEN, the restore should fail
	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "channel already open")
}

func (s *KeeperTestSuite) TestRestoreOracleICA_Failure_RegisterAccountFailure() {
	tc := s.SetupTestRestoreOracleICA()

	// Change the channel status to INIT so that it's not OPEN or CLOSED
	s.UpdateChannelState(tc.Oracle.PortId, tc.Oracle.ChannelId, channeltypes.INIT)

	// Disable middleware so the ICA registration fails
	s.App.ICAControllerKeeper.SetMiddlewareDisabled(s.Ctx, tc.Oracle.PortId, tc.Oracle.ConnectionId)

	_, err := s.GetMsgServer().RestoreOracleICA(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "unable to register oracle interchain account")
}
