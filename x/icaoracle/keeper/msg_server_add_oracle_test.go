package keeper_test

import (
	"fmt"

	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/Stride-Labs/stride/v30/x/icaoracle/types"
)

func (s *KeeperTestSuite) SetupTestAddOracle() types.MsgAddOracle {
	s.CreateTransferChannel(HostChainId)

	return types.MsgAddOracle{ConnectionId: ConnectionId}
}

func (s *KeeperTestSuite) TestAddOracle_Successful() {
	validMsg := s.SetupTestAddOracle()

	// Submit the AddOracle message
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().NoError(err, "no error expected when adding an oracle")

	// Confirm the oracle was created
	expectedOracle := types.Oracle{
		ChainId:      HostChainId,
		ConnectionId: ConnectionId,
		Active:       false,
	}
	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be created")
	s.Require().Equal(expectedOracle, actualOracle, "oracle created")

	// Confirm the ICA registration was initiated
	// We can verify this by checking that the ICAController module is bound to the oracle port
	expectedOraclePort := fmt.Sprintf("icacontroller-%s.ORACLE", HostChainId)
	_, isBound := s.App.ScopedICAControllerKeeper.GetCapability(s.Ctx, host.PortPath(expectedOraclePort))
	s.Require().True(isBound, "oracle ICA port %s should have been bound to the ICAController module", expectedOraclePort)
}

func (s *KeeperTestSuite) TestAddOracle_Successful_IcaAlreadyExists() {
	validMsg := s.SetupTestAddOracle()

	// Create the oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelID, portId := s.CreateICAChannel(owner)
	icaAddress := s.IcaAddresses[owner]

	// Submit the AddOracle message
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().NoError(err, "no error expected when adding an oracle")

	// Confirm the oracle was created and that the existing ICA channel was used
	expectedOracle := types.Oracle{
		ChainId:      HostChainId,
		ConnectionId: ConnectionId,
		ChannelId:    channelID,
		PortId:       portId,
		IcaAddress:   icaAddress,
		Active:       false,
	}
	actualOracle, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be created")
	s.Require().Equal(expectedOracle, actualOracle, "oracle created")
}

func (s *KeeperTestSuite) TestAddOracle_Failure_OracleAlreadyExists() {
	validMsg := s.SetupTestAddOracle()

	// Set an oracle successfully
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().NoError(err, "no error expected when adding an oracle")

	// Then attempt to submit it again - it should fail
	_, err = s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().ErrorContains(err, "oracle already exists", "error expected when adding duplicate oracle")
}

func (s *KeeperTestSuite) TestAddOracle_Failure_ControllerConnectionDoesNotExist() {
	validMsg := s.SetupTestAddOracle()

	// Submit the AddOracle message with an invalid connection Id - should fail
	invalidMsg := validMsg
	invalidMsg.ConnectionId = "fake_connection"
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "connection (fake_connection) not found", "error expected when adding oracle")
}

func (s *KeeperTestSuite) TestAddOracle_Failure_HostConnectionDoesNotExist() {
	validMsg := s.SetupTestAddOracle()

	// Delete the host connection ID from the controller connection end
	connectionEnd, found := s.App.IBCKeeper.ConnectionKeeper.GetConnection(s.Ctx, ConnectionId)
	s.Require().True(found, "connection should have been found")
	connectionEnd.Counterparty.ConnectionId = ""
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, ConnectionId, connectionEnd)

	// Submit the AddOracle message - it should fail
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().ErrorContains(err, "host connection not found", "error expected when adding oracle")
}

func (s *KeeperTestSuite) TestAddOracle_Failure_ClientDoesNotExist() {
	validMsg := s.SetupTestAddOracle()

	// Update the connection end so that the client cannot be found
	connectionEnd, found := s.App.IBCKeeper.ConnectionKeeper.GetConnection(s.Ctx, ConnectionId)
	s.Require().True(found, "connection should have been found")
	connectionEnd.ClientId = "fake_client"
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, ConnectionId, connectionEnd)

	// Submit the AddOracle message - it should fail
	_, err := s.GetMsgServer().AddOracle(s.Ctx, &validMsg)
	s.Require().ErrorContains(err, "client (fake_client) not found", "error expected when adding oracle")
}
