package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RestoreInterchainAccountTestCase struct {
	validMsg stakeibc.MsgRestoreInterchainAccount
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount() RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	hostZone := stakeibc.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		RedemptionRate: sdk.OneDec(), // if not yet, the beginblocker invariant panics
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	defaultMsg := stakeibc.MsgRestoreInterchainAccount{
		Creator:     "creatoraddress",
		ChainId:     HostChainId,
		AccountType: stakeibc.ICAAccountType_DELEGATION,
	}

	return RestoreInterchainAccountTestCase{
		validMsg: defaultMsg,
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount()
	owner := "GAIA.DELEGATION"
	channelID := s.CreateICAChannel(owner)
	portID := icatypes.PortPrefix + owner

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + delegate)")

	// Close the delegation channel
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx(), portID, channelID)
	s.Require().True(found, "delegation channel found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx(), portID, channelID, channel)

	// Restore the channel
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	s.Require().NoError(err, "registered ica account successfully")

	// Confirm the new channel was created
	channels = s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	s.Require().Len(channels, 3, "there should be 3 channels after restoring")

	// Confirm the new channel is in state INIT
	newChannelActive := false
	for _, channel := range channels {
		// The new channel should have the same port, a new channel ID and be in state INIT
		if channel.PortId == portID && channel.ChannelId != channelID && channel.State == channeltypes.INIT {
			newChannelActive = true
		}
	}
	s.Require().True(newChannelActive, "a new channel should have been created")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.AccountType = stakeibc.ICAAccountType_WITHDRAWAL
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMSg := fmt.Sprintf("ICA controller account address not found: %s.WITHDRAWAL: invalid interchain account address",
		tc.validMsg.ChainId)
	s.Require().EqualError(err, expectedErrMSg, "registered ica account successfully")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsForIncorrectHostZone() {
	tc := s.SetupRestoreInterchainAccount()
	msg := tc.validMsg
	msg.ChainId = "incorrectchainid"
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := "host zone not registered"
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails for incorrect host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_FailsIfAccountExists() {
	tc := s.SetupRestoreInterchainAccount()
	s.CreateICAChannel("GAIA.DELEGATION")
	msg := tc.validMsg
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx()), &msg)
	expectedErrMsg := fmt.Sprintf("existing active channel channel-1 for portID icacontroller-%s.DELEGATION on connection %s for owner %s.DELEGATION: active channel already set for this owner",
		tc.validMsg.ChainId,
		s.TransferPath.EndpointB.ConnectionID,
		tc.validMsg.ChainId,
	)
	s.Require().EqualError(err, expectedErrMsg, "registered ica account fails when account already exists")
}
