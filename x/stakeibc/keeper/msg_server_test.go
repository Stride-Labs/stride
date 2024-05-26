package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/utils"
	epochtypes "github.com/Stride-Labs/stride/v22/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v22/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v22/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v22/x/records/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

// ----------------------------------------------------
//	               RegisterHostZone
// ----------------------------------------------------

type RegisterHostZoneTestCase struct {
	validMsg                   stakeibctypes.MsgRegisterHostZone
	epochUnbondingRecordNumber uint64
	strideEpochNumber          uint64
	unbondingPeriod            uint64
	defaultRedemptionRate      sdk.Dec
	atomHostZoneChainId        string
}

func (s *KeeperTestSuite) SetupRegisterHostZone() RegisterHostZoneTestCase {
	epochUnbondingRecordNumber := uint64(3)
	strideEpochNumber := uint64(4)
	unbondingPeriod := uint64(14)
	defaultRedemptionRate := sdk.NewDec(1)
	atomHostZoneChainId := "GAIA"

	s.CreateTransferChannel(HostChainId)

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     epochUnbondingRecordNumber,
	})

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     strideEpochNumber,
	})

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochUnbondingRecordNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	defaultMsg := stakeibctypes.MsgRegisterHostZone{
		ConnectionId:      ibctesting.FirstConnectionID,
		Bech32Prefix:      GaiaPrefix,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		TransferChannelId: ibctesting.FirstChannelID,
		UnbondingPeriod:   unbondingPeriod,
		MinRedemptionRate: sdk.NewDec(0),
		MaxRedemptionRate: sdk.NewDec(0),
	}

	return RegisterHostZoneTestCase{
		validMsg:                   defaultMsg,
		epochUnbondingRecordNumber: epochUnbondingRecordNumber,
		strideEpochNumber:          strideEpochNumber,
		unbondingPeriod:            unbondingPeriod,
		defaultRedemptionRate:      defaultRedemptionRate,
		atomHostZoneChainId:        atomHostZoneChainId,
	}
}

// Helper function to test registering a duplicate host zone
// If there's a duplicate connection ID, register_host_zone will error before checking other fields for duplicates
// In order to test those cases, we need to first create a new host zone,
//
//	and then attempt to register with duplicate fields in the message
//
// This function 1) creates a new host zone and 2) returns what would be a successful register message
func (s *KeeperTestSuite) createNewHostZoneMessage(chainID string, denom string, prefix string) stakeibctypes.MsgRegisterHostZone {
	// Create a new test chain and connection ID
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp
	osmoChain := ibctesting.NewTestChain(s.T(), s.Coordinator, chainID)
	path := ibctesting.NewPath(s.StrideChain, osmoChain)
	s.Coordinator.SetupConnections(path)
	connectionId := path.EndpointA.ConnectionID

	// Build what would be a successful message to register the host zone
	// Note: this is purposefully missing fields because it is used in failure cases that short circuit
	return stakeibctypes.MsgRegisterHostZone{
		ConnectionId: connectionId,
		Bech32Prefix: prefix,
		HostDenom:    denom,
	}
}

// Helper function to assist in testing a failure to create an ICA account
// This function will occupy one of the specified port with the specified channel
//
//	so that the registration fails
func (s *KeeperTestSuite) createActiveChannelOnICAPort(accountName string, channelID string) {
	portID := fmt.Sprintf("%s%s.%s", icatypes.ControllerPortPrefix, HostChainId, accountName)
	openChannel := channeltypes.Channel{State: channeltypes.OPEN}

	// The channel ID doesn't matter here - all that matters is that theres an open channel on the port
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portID, channelID, openChannel)
	s.App.ICAControllerKeeper.SetActiveChannelID(s.Ctx, ibctesting.FirstConnectionID, portID, channelID)
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success() {
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// Register host zone
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "able to successfully register host zone")

	// Confirm host zone unbonding was added
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(tc.defaultRedemptionRate, hostZone.RedemptionRate, "redemption rate set to default: 1")
	s.Require().Equal(tc.defaultRedemptionRate, hostZone.LastRedemptionRate, "last redemption rate set to default: 1")
	defaultMinThreshold := sdk.NewDec(int64(stakeibctypes.DefaultMinRedemptionRateThreshold)).Quo(sdk.NewDec(100))
	defaultMaxThreshold := sdk.NewDec(int64(stakeibctypes.DefaultMaxRedemptionRateThreshold)).Quo(sdk.NewDec(100))
	s.Require().Equal(defaultMinThreshold, hostZone.MinRedemptionRate, "min redemption rate set to default")
	s.Require().Equal(defaultMaxThreshold, hostZone.MaxRedemptionRate, "max redemption rate set to default")
	s.Require().Equal(tc.unbondingPeriod, hostZone.UnbondingPeriod, "unbonding period")

	// Confirm host zone unbonding record was created
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.epochUnbondingRecordNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Len(epochUnbondingRecord.HostZoneUnbondings, 1, "host zone unbonding record has one entry")

	// Confirm host zone unbonding was added
	hostZoneUnbonding := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(HostChainId, hostZoneUnbonding.HostZoneId, "host zone unbonding set for this host zone")
	s.Require().Equal(sdkmath.ZeroInt(), hostZoneUnbonding.NativeTokenAmount, "host zone unbonding set to 0 tokens")
	s.Require().Equal(recordstypes.HostZoneUnbonding_UNBONDING_QUEUE, hostZoneUnbonding.Status, "host zone unbonding set to bonded")

	// Confirm a module account was created
	hostZoneModuleAccount, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	s.Require().NoError(err, "converting module address to account")
	acc := s.App.AccountKeeper.GetAccount(s.Ctx, hostZoneModuleAccount)
	s.Require().NotNil(acc, "host zone module account found in account keeper")

	// Confirm an empty deposit record was created
	expectedDepositRecord := recordstypes.DepositRecord{
		Id:                 uint64(0),
		Amount:             sdkmath.ZeroInt(),
		HostZoneId:         hostZone.ChainId,
		Denom:              hostZone.HostDenom,
		Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
		DepositEpochNumber: tc.strideEpochNumber,
	}

	depositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(depositRecords, 1, "number of deposit records")
	s.Require().Equal(expectedDepositRecord, depositRecords[0], "deposit record")

	// Confirm max ICA messages was set to default
	s.Require().Equal(keeper.DefaultMaxMessagesPerIcaTx, hostZone.MaxMessagesPerIcaTx, "max messages per ica tx")
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success_SetCommunityPoolTreasuryAddress() {
	tc := s.SetupRegisterHostZone()

	// Sets the community pool treasury address to a valid address
	msg := tc.validMsg
	msg.CommunityPoolTreasuryAddress = ValidHostAddress

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when registering host with valid treasury address")

	// Confirm treasury address was set
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(ValidHostAddress, hostZone.CommunityPoolTreasuryAddress, "treasury address")
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success_SetMaxIcaMessagesPerTx() {
	tc := s.SetupRegisterHostZone()

	// Set the max number of ICA messages
	maxMessages := uint64(100)
	msg := tc.validMsg
	msg.MaxMessagesPerIcaTx = maxMessages

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when registering host with max messages")

	// Confirm max number of messages was set
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(maxMessages, hostZone.MaxMessagesPerIcaTx, "max messages per ica tx")
}

func (s *KeeperTestSuite) TestRegisterHostZone_Success_Unregister() {
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// Register the host zone with the valid message
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when registering host")

	// Confirm accounts were created
	depositAddress := types.NewHostZoneDepositAddress(chainId)
	communityPoolStakeAddress := types.NewHostZoneModuleAddress(chainId, keeper.CommunityPoolStakeHoldingAddressKey)
	communityPoolRedeemAddress := types.NewHostZoneModuleAddress(chainId, keeper.CommunityPoolRedeemHoldingAddressKey)

	depositAccount := s.App.AccountKeeper.GetAccount(s.Ctx, depositAddress)
	communityPoolStakeAccount := s.App.AccountKeeper.GetAccount(s.Ctx, communityPoolStakeAddress)
	communityPoolRedeemAccount := s.App.AccountKeeper.GetAccount(s.Ctx, communityPoolRedeemAddress)

	s.Require().NotNil(depositAccount, "deposit account should exist")
	s.Require().NotNil(communityPoolStakeAccount, "community pool stake account should exist")
	s.Require().NotNil(communityPoolRedeemAccount, "community pool redeem account should exist")

	// Confirm records were created
	depositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(depositRecords, 1, "there should be one deposit record")

	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().Len(epochUnbondingRecords, 1, "there should be one epoch unbonding record")
	s.Require().Len(epochUnbondingRecords[0].HostZoneUnbondings, 1, "there should be one host zone unbonding record")

	// Unregister the host zone
	err = s.App.StakeibcKeeper.UnregisterHostZone(s.Ctx, HostChainId)
	s.Require().NoError(err, "no error expected when unregistering host zone")

	// Confirm accounts were deleted
	depositAccount = s.App.AccountKeeper.GetAccount(s.Ctx, depositAddress)
	communityPoolStakeAccount = s.App.AccountKeeper.GetAccount(s.Ctx, communityPoolStakeAddress)
	communityPoolRedeemAccount = s.App.AccountKeeper.GetAccount(s.Ctx, communityPoolRedeemAddress)

	s.Require().Nil(depositAccount, "deposit account should have been deleted")
	s.Require().Nil(communityPoolStakeAccount, "community pool stake account should have been deleted")
	s.Require().Nil(communityPoolRedeemAccount, "community pool redeem account should have been deleted")

	// Confirm records were deleted
	depositRecords = s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Empty(depositRecords, "deposit records should have been deleted")

	epochUnbondingRecords = s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().Empty(epochUnbondingRecords[0].HostZoneUnbondings, "host zone unbonding record should have been deleted")

	// Attempt to re-register, it should succeed
	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when re-registering host")
}

func (s *KeeperTestSuite) TestRegisterHostZone_InvalidConnectionId() {
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg
	msg.ConnectionId = "connection-10" // an invalid connection ID

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().EqualError(err, "invalid connection id, connection-10 not found: failed to register host zone")
}

func (s *KeeperTestSuite) TestRegisterHostZone_DuplicateConnectionIdInIBCState() {
	// tests for a failure if we register the same host zone twice
	// (with a duplicate connectionId stored in the IBCKeeper's state)
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "able to successfully register host zone once")

	// now all attributes are different, EXCEPT the connection ID
	msg.Bech32Prefix = "cosmos-different" // a different Bech32 prefix
	msg.HostDenom = "atom-different"      // a different host denom
	msg.IbcDenom = "ibc-atom-different"   // a different IBC denom

	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := "invalid chain id, zone for GAIA already registered: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "registering host zone with duplicate connection ID should fail")
}

func (s *KeeperTestSuite) TestRegisterHostZone_DuplicateConnectionIdInStakeibcState() {
	// tests for a failure if we register the same host zone twice
	// (with a duplicate connectionId stored in a different host zone in stakeibc)
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "able to successfully register host zone once")

	// Create the message for a brand new host zone
	// (without modifications, you would expect this to be successful)
	newHostZoneMsg := s.createNewHostZoneMessage("OSMO", "osmo", "osmo")

	// Add a different host zone with the same connection Id as OSMO
	newHostZone := stakeibctypes.HostZone{
		ChainId:      "JUNO",
		ConnectionId: newHostZoneMsg.ConnectionId,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, newHostZone)

	// Registering should fail with a duplicate connection ID
	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &newHostZoneMsg)
	expectedErrMsg := "connectionId connection-1 already registered: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "registering host zone with duplicate connection ID should fail")
}

func (s *KeeperTestSuite) TestRegisterHostZone_DuplicateHostDenom() {
	// tests for a failure if we register the same host zone twice (with a duplicate host denom)
	tc := s.SetupRegisterHostZone()

	// Register host zones successfully
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "able to successfully register host zone once")

	// Create the message for a brand new host zone
	// (without modifications, you would expect this to be successful)
	newHostZoneMsg := s.createNewHostZoneMessage("OSMO", "osmo", "osmo")

	// Try to register with a duplicate host denom - it should fail
	invalidMsg := newHostZoneMsg
	invalidMsg.HostDenom = tc.validMsg.HostDenom

	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	expectedErrMsg := "host denom uatom already registered: failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "registering host zone with duplicate host denom should fail")
}

func (s *KeeperTestSuite) TestRegisterHostZone_DuplicateTransferChannel() {
	// tests for a failure if we register the same host zone twice (with a duplicate transfer)
	tc := s.SetupRegisterHostZone()

	// Register host zones successfully
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "able to successfully register host zone once")

	// Create the message for a brand new host zone
	// (without modifications, you would expect this to be successful)
	newHostZoneMsg := s.createNewHostZoneMessage("OSMO", "osmo", "osmo")

	// Try to register with a duplicate transfer channel - it should fail
	invalidMsg := newHostZoneMsg
	invalidMsg.TransferChannelId = tc.validMsg.TransferChannelId

	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	expectedErrMsg := "transfer channel channel-0 already registered: failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "registering host zone with duplicate host denom should fail")
}

func (s *KeeperTestSuite) TestRegisterHostZone_DuplicateBech32Prefix() {
	// tests for a failure if we register the same host zone twice (with a duplicate bech32 prefix)
	tc := s.SetupRegisterHostZone()

	// Register host zones successfully
	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "able to successfully register host zone once")

	// Create the message for a brand new host zone
	// (without modifications, you would expect this to be successful)
	newHostZoneMsg := s.createNewHostZoneMessage("OSMO", "osmo", "osmo")

	// Try to register with a duplicate bech32prefix - it should fail
	invalidMsg := newHostZoneMsg
	invalidMsg.Bech32Prefix = tc.validMsg.Bech32Prefix

	_, err = s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	expectedErrMsg := "bech32prefix cosmos already registered: failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "registering host zone with duplicate bech32 prefix should fail")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotFindDayEpochTracker() {
	// tests for a failure if the epoch tracker cannot be found
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// delete the epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.DAY_EPOCH)

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := "epoch tracker (day) not found: epoch not found"
	s.Require().EqualError(err, expectedErrMsg, "day epoch tracker not found")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotFindStrideEpochTracker() {
	// tests for a failure if the epoch tracker cannot be found
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// delete the epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := "epoch tracker (stride_epoch) not found: epoch not found"
	s.Require().EqualError(err, expectedErrMsg, "stride epoch tracker not found")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotFindEpochUnbondingRecord() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()
	msg := tc.validMsg

	// delete the epoch unbonding record
	s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, tc.epochUnbondingRecordNumber)

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &msg)
	expectedErrMsg := "unable to find latest epoch unbonding record: epoch unbonding record not found"
	s.Require().EqualError(err, expectedErrMsg, " epoch unbonding record not found")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterDelegationAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on delegation port
	s.createActiveChannelOnICAPort("DELEGATION", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErrMsg := "unable to register delegation account, err: existing active channel channel-1 for portID icacontroller-GAIA.DELEGATION "
	expectedErrMsg += "on connection connection-0: active channel already set for this owner: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "can't register delegation account")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterFeeAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on fee port
	s.createActiveChannelOnICAPort("FEE", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErrMsg := "unable to register fee account, err: existing active channel channel-1 for portID icacontroller-GAIA.FEE "
	expectedErrMsg += "on connection connection-0: active channel already set for this owner: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "can't register redemption account")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterWithdrawalAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on withdrawal port
	s.createActiveChannelOnICAPort("WITHDRAWAL", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErrMsg := "unable to register withdrawal account, err: existing active channel channel-1 for portID icacontroller-GAIA.WITHDRAWAL "
	expectedErrMsg += "on connection connection-0: active channel already set for this owner: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "can't register redemption account")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterRedemptionAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on redemption port
	s.createActiveChannelOnICAPort("REDEMPTION", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	expectedErrMsg := "unable to register redemption account, err: existing active channel channel-1 for portID icacontroller-GAIA.REDEMPTION "
	expectedErrMsg += "on connection connection-0: active channel already set for this owner: "
	expectedErrMsg += "failed to register host zone"
	s.Require().EqualError(err, expectedErrMsg, "can't register redemption account")
}

func (s *KeeperTestSuite) TestRegisterHostZone_InvalidCommunityPoolTreasuryAddress() {
	// tests for a failure if the community pool treasury address is invalid
	tc := s.SetupRegisterHostZone()

	invalidMsg := tc.validMsg
	invalidMsg.CommunityPoolTreasuryAddress = "invalid_address"

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid community pool treasury address")
}

// ----------------------------------------------------
//	             UpdateHostZoneParams
// ----------------------------------------------------

func (s *KeeperTestSuite) TestUpdateHostZoneParams() {
	initialMessages := uint64(32)
	updatedMessages := uint64(100)

	// Create a host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:             HostChainId,
		MaxMessagesPerIcaTx: initialMessages,
	})

	// Submit the message to update the params
	validUpdateMsg := types.MsgUpdateHostZoneParams{
		Authority:           Authority,
		ChainId:             HostChainId,
		MaxMessagesPerIcaTx: updatedMessages,
	}
	_, err := s.GetMsgServer().UpdateHostZoneParams(sdk.WrapSDKContext(s.Ctx), &validUpdateMsg)
	s.Require().NoError(err, "no error expected when updating host zone params")

	// Check that the max messages was updated
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(updatedMessages, hostZone.MaxMessagesPerIcaTx, "max messages")

	// Update it again, setting it to the default value
	validUpdateMsg = types.MsgUpdateHostZoneParams{
		Authority:           Authority,
		ChainId:             HostChainId,
		MaxMessagesPerIcaTx: 0,
	}
	_, err = s.GetMsgServer().UpdateHostZoneParams(sdk.WrapSDKContext(s.Ctx), &validUpdateMsg)
	s.Require().NoError(err, "no error expected when updating host zone params again")

	// Check that the max messages was updated
	hostZone = s.MustGetHostZone(HostChainId)
	s.Require().Equal(keeper.DefaultMaxMessagesPerIcaTx, hostZone.MaxMessagesPerIcaTx, "max messages")

	// Attempt it again with an invalid chain ID, it should fail
	invalidUpdateMsg := types.MsgUpdateHostZoneParams{
		Authority:           Authority,
		ChainId:             "missing-host",
		MaxMessagesPerIcaTx: updatedMessages,
	}
	_, err = s.GetMsgServer().UpdateHostZoneParams(sdk.WrapSDKContext(s.Ctx), &invalidUpdateMsg)
	s.Require().ErrorContains(err, "host zone not found")

	// Finally attempt again with an invalid authority, it should also fail
	invalidUpdateMsg = types.MsgUpdateHostZoneParams{
		Authority:           "invalid-authority",
		ChainId:             HostChainId,
		MaxMessagesPerIcaTx: updatedMessages,
	}
	_, err = s.GetMsgServer().UpdateHostZoneParams(sdk.WrapSDKContext(s.Ctx), &invalidUpdateMsg)
	s.Require().ErrorContains(err, "invalid authority")
}

// ----------------------------------------------------
//	                  AddValidator
// ----------------------------------------------------

type AddValidatorsTestCase struct {
	hostZone                 types.HostZone
	validMsg                 types.MsgAddValidators
	expectedValidators       []*types.Validator
	validatorQueryDataToName map[string]string
}

// Helper function to determine the validator's key in the staking store
// which is used as the request data in the ICQ
func (s *KeeperTestSuite) getSharesToTokensRateQueryData(validatorAddress string) []byte {
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	s.Require().NoError(err, "no error expected when decoding validator address")
	return stakingtypes.GetValidatorKey(validatorAddressBz)
}

func (s *KeeperTestSuite) SetupAddValidators() AddValidatorsTestCase {
	slashThreshold := uint64(10)
	params := types.DefaultParams()
	params.ValidatorSlashQueryThreshold = slashThreshold
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	totalDelegations := sdkmath.NewInt(100_000)
	expectedSlashCheckpoint := sdkmath.NewInt(10_000)

	hostZone := types.HostZone{
		ChainId:          "GAIA",
		ConnectionId:     ibctesting.FirstConnectionID,
		Validators:       []*types.Validator{},
		TotalDelegations: totalDelegations,
	}

	validatorAddresses := map[string]string{
		"val1": "stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm",
		"val2": "stridevaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rcfud5c",
		"val3": "stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f",
	}

	// mapping of query request data to validator name
	// serves as a reverse lookup to map sharesToTokens rate queries to validators
	validatorQueryDataToName := map[string]string{}
	for name, address := range validatorAddresses {
		queryData := s.getSharesToTokensRateQueryData(address)
		validatorQueryDataToName[string(queryData)] = name
	}

	validMsg := types.MsgAddValidators{
		Creator:  "stride_ADMIN",
		HostZone: HostChainId,
		Validators: []*types.Validator{
			{Name: "val1", Address: validatorAddresses["val1"], Weight: 1},
			{Name: "val2", Address: validatorAddresses["val2"], Weight: 2},
			{Name: "val3", Address: validatorAddresses["val3"], Weight: 3},
		},
	}

	expectedValidators := []*types.Validator{
		{Name: "val1", Address: validatorAddresses["val1"], Weight: 1},
		{Name: "val2", Address: validatorAddresses["val2"], Weight: 2},
		{Name: "val3", Address: validatorAddresses["val3"], Weight: 3},
	}
	for _, validator := range expectedValidators {
		validator.Delegation = sdkmath.ZeroInt()
		validator.SlashQueryProgressTracker = sdkmath.ZeroInt()
		validator.SharesToTokensRate = sdk.ZeroDec()
		validator.SlashQueryCheckpoint = expectedSlashCheckpoint
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the latest client height for the ICQ submission
	s.MockClientLatestHeight(1)

	return AddValidatorsTestCase{
		hostZone:                 hostZone,
		validMsg:                 validMsg,
		expectedValidators:       expectedValidators,
		validatorQueryDataToName: validatorQueryDataToName,
	}
}

func (s *KeeperTestSuite) TestAddValidators_Successful() {
	tc := s.SetupAddValidators()

	// Add validators
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(3, len(hostZone.Validators), "number of validators")

	for i := 0; i < 3; i++ {
		s.Require().Equal(*tc.expectedValidators[i], *hostZone.Validators[i], "validators %d", i)
	}

	// Confirm ICQs were submitted
	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 3)

	// Map the query responses to the validator names to get the names of the validators that
	// were queried
	queriedValidators := []string{}
	for i, query := range queries {
		validator, ok := tc.validatorQueryDataToName[string(query.RequestData)]
		s.Require().True(ok, "query from response %d does not match any expected requests", i)
		queriedValidators = append(queriedValidators, validator)
	}

	// Confirm the list of queried validators matches the full list of validators
	allValidatorNames := []string{}
	for _, expected := range tc.expectedValidators {
		allValidatorNames = append(allValidatorNames, expected.Name)
	}
	s.Require().ElementsMatch(allValidatorNames, queriedValidators, "queried validators")
}

func (s *KeeperTestSuite) TestAddValidators_HostZoneNotFound() {
	tc := s.SetupAddValidators()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsg
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	s.Require().EqualError(err, "Host Zone (gaia) not found: host zone not found")
}

func (s *KeeperTestSuite) TestAddValidators_AddressAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that the name val1 already exists
	hostZone := tc.hostZone
	duplicateAddress := tc.expectedValidators[0].Address
	duplicateVal := types.Validator{Name: "new_val", Address: duplicateAddress}
	hostZone.Validators = []*types.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator address to val1 so that the message errors
	expectedError := fmt.Sprintf("Validator address (%s) already exists on Host Zone (GAIA)", duplicateAddress)
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, expectedError)
}

func (s *KeeperTestSuite) TestAddValidators_NameAlreadyExists() {
	tc := s.SetupAddValidators()

	// Update host zone so that val1's address already exists
	hostZone := tc.hostZone
	duplicateName := tc.expectedValidators[0].Name
	duplicateVal := types.Validator{Name: duplicateName, Address: "new_address"}
	hostZone.Validators = []*types.Validator{&duplicateVal}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Change the validator name to val1 so that the message errors
	expectedError := fmt.Sprintf("Validator name (%s) already exists on Host Zone (GAIA)", duplicateName)
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, expectedError)
}

func (s *KeeperTestSuite) TestAddValidators_SuccessfulManyValidators() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
	})
	s.MockClientLatestHeight(1)

	// Setup validators in a top-heavy order so that *if* the weight cap
	// was checked after each validator, it would fail midway
	// However, the addition of last validator causes the highest weight
	// validator to be below 10%
	validators := []*types.Validator{
		{Name: "val1", Weight: 10},
		{Name: "val2", Weight: 10},
		{Name: "val3", Weight: 9},
		{Name: "val4", Weight: 9},
		{Name: "val5", Weight: 8},
		{Name: "val6", Weight: 8},
		{Name: "val7", Weight: 7},
		{Name: "val8", Weight: 7},
		{Name: "val9", Weight: 6},
		{Name: "val10", Weight: 6},
		{Name: "val11", Weight: 5},
		{Name: "val12", Weight: 5},
		{Name: "val13", Weight: 4},
		{Name: "val14", Weight: 4},
		{Name: "val15", Weight: 3},
	}

	// Assign an address for each
	addresses := apptesting.CreateRandomAccounts(len(validators))
	for i, validator := range validators {
		validator.Address = addresses[i].String()
	}

	// Submit the add validator message - it should succeed
	addValidatorMsg := types.MsgAddValidators{
		HostZone:   HostChainId,
		Validators: validators,
	}
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &addValidatorMsg)
	s.Require().NoError(err, "no error expected when adding validators")
}

func (s *KeeperTestSuite) TestAddValidators_ValidatorWeightCapExceeded() {
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
	})
	s.MockClientLatestHeight(1)

	// The distribution below will lead to the first two validators owning more
	// than a 10% share
	validators := []*types.Validator{
		{Name: "val1", Weight: 10},
		{Name: "val2", Weight: 10},
		{Name: "val3", Weight: 9},
		{Name: "val4", Weight: 9},
		{Name: "val5", Weight: 8},
		{Name: "val6", Weight: 8},
		{Name: "val7", Weight: 7},
		{Name: "val8", Weight: 7},
		{Name: "val9", Weight: 6},
		{Name: "val10", Weight: 6},
		{Name: "val11", Weight: 5},
		{Name: "val12", Weight: 5},
		{Name: "val13", Weight: 4},
		{Name: "val14", Weight: 4},
	}

	// Assign an address for each
	addresses := apptesting.CreateRandomAccounts(len(validators))
	for i, validator := range validators {
		validator.Address = addresses[i].String()
	}

	// Submit the add validator message - it should error
	addValidatorMsg := types.MsgAddValidators{
		HostZone:   HostChainId,
		Validators: validators,
	}
	_, err := s.GetMsgServer().AddValidators(sdk.WrapSDKContext(s.Ctx), &addValidatorMsg)
	s.Require().ErrorContains(err, "validator exceeds weight cap")
}

// ----------------------------------------------------
//	               DeleteValidator
// ----------------------------------------------------

type DeleteValidatorTestCase struct {
	hostZone          stakeibctypes.HostZone
	initialValidators []*stakeibctypes.Validator
	validMsgs         []stakeibctypes.MsgDeleteValidator
}

func (s *KeeperTestSuite) SetupDeleteValidator() DeleteValidatorTestCase {
	initialValidators := []*stakeibctypes.Validator{
		{
			Name:               "val1",
			Address:            "stride_VAL1",
			Weight:             0,
			Delegation:         sdkmath.ZeroInt(),
			SharesToTokensRate: sdk.OneDec(),
		},
		{
			Name:               "val2",
			Address:            "stride_VAL2",
			Weight:             0,
			Delegation:         sdkmath.ZeroInt(),
			SharesToTokensRate: sdk.OneDec(),
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:    "GAIA",
		Validators: initialValidators,
	}
	validMsgs := []stakeibctypes.MsgDeleteValidator{
		{
			Creator:  "stride_ADDRESS",
			HostZone: "GAIA",
			ValAddr:  "stride_VAL1",
		},
		{
			Creator:  "stride_ADDRESS",
			HostZone: "GAIA",
			ValAddr:  "stride_VAL2",
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return DeleteValidatorTestCase{
		hostZone:          hostZone,
		initialValidators: initialValidators,
		validMsgs:         validMsgs,
	}
}

func (s *KeeperTestSuite) TestDeleteValidator_Successful() {
	tc := s.SetupDeleteValidator()

	// Delete first validator
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(1, len(hostZone.Validators), "number of validators should be 1")
	s.Require().Equal(tc.initialValidators[1:], hostZone.Validators, "validators list after removing 1 validator")

	// Delete second validator
	_, err = s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[1])
	s.Require().NoError(err)

	hostZone, found = s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(0, len(hostZone.Validators), "number of validators should be 0")
}

func (s *KeeperTestSuite) TestDeleteValidator_HostZoneNotFound() {
	tc := s.SetupDeleteValidator()

	// Replace hostzone in msg to a host zone that doesn't exist
	badHostZoneMsg := tc.validMsgs[0]
	badHostZoneMsg.HostZone = "gaia"
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &badHostZoneMsg)
	errMsg := "Validator (stride_VAL1) not removed from host zone (gaia) "
	errMsg += "| err: HostZone (gaia) not found: host zone not found: validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_AddressNotFound() {
	tc := s.SetupDeleteValidator()

	// Build message with a validator address that does not exist
	badAddressMsg := tc.validMsgs[0]
	badAddressMsg.ValAddr = "stride_VAL5"
	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &badAddressMsg)

	errMsg := "Validator (stride_VAL5) not removed from host zone (GAIA) "
	errMsg += "| err: Validator address (stride_VAL5) not found on host zone (GAIA): "
	errMsg += "validator not found: validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroDelegation() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero delegation
	hostZone := tc.hostZone
	hostZone.Validators[0].Delegation = sdkmath.NewInt(1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	errMsg := "Validator (stride_VAL1) not removed from host zone (GAIA) "
	errMsg += "| err: Validator (stride_VAL1) has non-zero delegation (1) or weight (0): "
	errMsg += "validator not removed"
	s.Require().EqualError(err, errMsg)
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroWeight() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero weight
	hostZone := tc.hostZone
	hostZone.Validators[0].Weight = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(sdk.WrapSDKContext(s.Ctx), &tc.validMsgs[0])
	errMsg := "Validator (stride_VAL1) not removed from host zone (GAIA) "
	errMsg += "| err: Validator (stride_VAL1) has non-zero delegation (0) or weight (1): "
	errMsg += "validator not removed"
	s.Require().EqualError(err, errMsg)
}

// ----------------------------------------------------
//	                 ClearBalance
// ----------------------------------------------------

type ClearBalanceState struct {
	feeChannel Channel
	hz         stakeibctypes.HostZone
}

type ClearBalanceTestCase struct {
	initialState ClearBalanceState
	validMsg     stakeibctypes.MsgClearBalance
}

func (s *KeeperTestSuite) SetupClearBalance() ClearBalanceTestCase {
	// fee account
	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	feeChannelID, _ := s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]
	// hz
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)
	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		DepositAddress: depositAddress.String(),
		FeeIcaAddress:  feeAddress,
	}

	amount := sdkmath.NewInt(1_000_000)

	user := Account{
		acc: s.TestAccs[0],
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return ClearBalanceTestCase{
		initialState: ClearBalanceState{
			hz: hostZone,
			feeChannel: Channel{
				PortID:    icatypes.ControllerPortPrefix + feeAccountOwner,
				ChannelID: feeChannelID,
			},
		},
		validMsg: stakeibctypes.MsgClearBalance{
			Creator: user.acc.String(),
			ChainId: HostChainId,
			Amount:  amount,
			Channel: feeChannelID,
		},
	}
}

func (s *KeeperTestSuite) TestClearBalance_Successful() {
	tc := s.SetupClearBalance()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	feeChannel := tc.initialState.feeChannel
	feePortId := feeChannel.PortID
	feeChannelId := feeChannel.ChannelID

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found before clear balance")

	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "balance clears")

	// Confirm the sequence number was incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found after clear balance")
	s.Require().Equal(endSequence, startSequence+1, "sequence number after clear balance")
}

func (s *KeeperTestSuite) TestClearBalance_HostChainMissing() {
	tc := s.SetupClearBalance()
	// remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "chainId: GAIA: host zone not registered")
}

func (s *KeeperTestSuite) TestClearBalance_FeeAccountMissing() {
	tc := s.SetupClearBalance()
	// no fee account
	tc.initialState.hz.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.initialState.hz)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "fee acount not found for chainId: GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestClearBalance_ParseCoinError() {
	tc := s.SetupClearBalance()
	// invalid denom
	tc.initialState.hz.HostDenom = ":"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.initialState.hz)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "failed to parse coin (1000000:): invalid decimal coin expression: 1000000:")
}

// ----------------------------------------------------
//	                 LiquidStake
// ----------------------------------------------------

type Account struct {
	acc           sdk.AccAddress
	atomBalance   sdk.Coin
	stAtomBalance sdk.Coin
}

type LiquidStakeState struct {
	depositRecordAmount sdkmath.Int
	hostZone            stakeibctypes.HostZone
}

type LiquidStakeTestCase struct {
	user         Account
	zoneAccount  Account
	initialState LiquidStakeState
	validMsg     stakeibctypes.MsgLiquidStake
}

func (s *KeeperTestSuite) SetupLiquidStake() LiquidStakeTestCase {
	stakeAmount := sdkmath.NewInt(1_000_000)
	initialDepositAmount := sdkmath.NewInt(1_000_000)
	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 0),
	}
	s.FundAccount(user.acc, user.atomBalance)

	depositAddress := stakeibctypes.NewHostZoneDepositAddress(HostChainId)

	zoneAccount := Account{
		acc:           depositAddress,
		atomBalance:   sdk.NewInt64Coin(IbcAtom, 10_000_000),
		stAtomBalance: sdk.NewInt64Coin(StAtom, 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		DepositAddress: depositAddress.String(),
	}

	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	}

	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         "GAIA",
		Amount:             initialDepositAmount,
		Status:             recordtypes.DepositRecord_TRANSFER_QUEUE,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord)

	return LiquidStakeTestCase{
		user:        user,
		zoneAccount: zoneAccount,
		initialState: LiquidStakeState{
			depositRecordAmount: initialDepositAmount,
			hostZone:            hostZone,
		},
		validMsg: stakeibctypes.MsgLiquidStake{
			Creator:   user.acc.String(),
			HostDenom: Atom,
			Amount:    stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestLiquidStake_Successful() {
	tc := s.SetupLiquidStake()
	user := tc.user
	zoneAccount := tc.zoneAccount
	msg := tc.validMsg
	initialStAtomSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err)

	// Confirm balances
	// User IBC/UATOM balance should have DECREASED by the size of the stake
	expectedUserAtomBalance := user.atomBalance.SubAmount(msg.Amount)
	actualUserAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, IbcAtom)
	// zoneAccount IBC/UATOM balance should have INCREASED by the size of the stake
	expectedzoneAccountAtomBalance := zoneAccount.atomBalance.AddAmount(msg.Amount)
	actualzoneAccountAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount.acc, IbcAtom)
	// User STUATOM balance should have INCREASED by the size of the stake
	expectedUserStAtomBalance := user.stAtomBalance.AddAmount(msg.Amount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom)
	// Bank supply of STUATOM should have INCREASED by the size of the stake
	expectedBankSupply := initialStAtomSupply.AddAmount(msg.Amount)
	actualBankSupply := s.App.BankKeeper.GetSupply(s.Ctx, StAtom)

	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")
	s.CompareCoins(expectedUserAtomBalance, actualUserAtomBalance, "user ibc/uatom balance")
	s.CompareCoins(expectedzoneAccountAtomBalance, actualzoneAccountAtomBalance, "zoneAccount ibc/uatom balance")
	s.CompareCoins(expectedBankSupply, actualBankSupply, "bank stuatom supply")

	// Confirm deposit record adjustment
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")

	expectedDepositRecordAmount := tc.initialState.depositRecordAmount.Add(msg.Amount)
	actualDepositRecordAmount := records[0].Amount
	s.Require().Equal(expectedDepositRecordAmount, actualDepositRecordAmount, "deposit record amount")
}

func (s *KeeperTestSuite) TestLiquidStake_DifferentRedemptionRates() {
	tc := s.SetupLiquidStake()
	user := tc.user
	msg := tc.validMsg

	// Loop over sharesToTokens rates: {0.92, 0.94, ..., 1.2}
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := sdk.NewDecWithPrec(1.0, 1).Quo(sdk.NewDec(10)).Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.02
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate

		// Update rate in host zone
		hz := tc.initialState.hostZone
		hz.RedemptionRate = newRedemptionRate
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom).Amount
		_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
		s.Require().NoError(err)
		endingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, StAtom).Amount
		actualStAtomMinted := endingStAtomBalance.Sub(startingStAtomBalance)

		expectedStAtomMinted := sdk.NewDecFromInt(msg.Amount).Quo(redemptionRateFloat).TruncateInt()
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		s.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)
	}
}

func (s *KeeperTestSuite) TestLiquidStake_HostZoneNotFound() {
	tc := s.SetupLiquidStake()
	// Update message with invalid denom
	invalidMsg := tc.validMsg
	invalidMsg.HostDenom = "ufakedenom"
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "no host zone found for denom (ufakedenom): invalid token denom")
}

func (s *KeeperTestSuite) TestLiquidStake_HostZoneHalted() {
	tc := s.SetupLiquidStake()

	// Update the host zone so that it's halted
	badHostZone := tc.initialState.hostZone
	badHostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "halted host zone found for denom (uatom): Halted host zone found")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidUserAddress() {
	tc := s.SetupLiquidStake()

	// Update hostzone with invalid address
	invalidMsg := tc.validMsg
	invalidMsg.Creator = "cosmosXXX"

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "user's address is invalid: decoding bech32 failed: string not all lowercase or all uppercase")
}

func (s *KeeperTestSuite) TestLiquidStake_InvalidHostAddress() {
	tc := s.SetupLiquidStake()

	// Update hostzone with invalid address
	badHostZone := tc.initialState.hostZone
	badHostZone.DepositAddress = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "host zone address is invalid: decoding bech32 failed: string not all lowercase or all uppercase")
}

func (s *KeeperTestSuite) TestLiquidStake_RateBelowMinThreshold() {
	tc := s.SetupLiquidStake()
	msg := tc.validMsg

	// Update rate in host zone to below min threshold
	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.MustNewDecFromStr("0.8")
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestLiquidStake_RateAboveMaxThreshold() {
	tc := s.SetupLiquidStake()
	msg := tc.validMsg

	// Update rate in host zone to below min threshold
	hz := tc.initialState.hostZone
	hz.RedemptionRate = sdk.NewDec(2)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestLiquidStake_NoEpochTracker() {
	tc := s.SetupLiquidStake()
	// Remove epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no epoch number for epoch (%s): not found", epochtypes.STRIDE_EPOCH))
}

func (s *KeeperTestSuite) TestLiquidStake_NoDepositRecord() {
	tc := s.SetupLiquidStake()
	// Remove deposit record
	s.App.RecordsKeeper.RemoveDepositRecord(s.Ctx, 1)
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("no deposit record for epoch (%d): not found", 1))
}

func (s *KeeperTestSuite) TestLiquidStake_NotIbcDenom() {
	tc := s.SetupLiquidStake()
	// Update hostzone with non-ibc denom
	badDenom := "i/uatom"
	badHostZone := tc.initialState.hostZone
	badHostZone.IbcDenom = badDenom
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)
	// Fund the user with the non-ibc denom
	s.FundAccount(tc.user.acc, sdk.NewInt64Coin(badDenom, 1000000000))
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("denom is not an IBC token (%s): invalid token denom", badHostZone.IbcDenom))
}

func (s *KeeperTestSuite) TestLiquidStake_ZeroStTokens() {
	tc := s.SetupLiquidStake()

	// Adjust redemption rate and liquid stake amount so that the number of stTokens would be zero
	// stTokens = 1(amount) / 1.1(RR) = rounds down to 0
	hostZone := tc.initialState.hostZone
	hostZone.RedemptionRate = sdk.NewDecWithPrec(11, 1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	tc.validMsg.Amount = sdkmath.NewInt(1)

	// The liquid stake should fail
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "Liquid stake of 1uatom would return 0 stTokens: Liquid staked amount is too small")
}

func (s *KeeperTestSuite) TestLiquidStake_InsufficientBalance() {
	tc := s.SetupLiquidStake()
	// Set liquid stake amount to value greater than account balance
	invalidMsg := tc.validMsg
	balance := tc.user.atomBalance.Amount
	invalidMsg.Amount = balance.Add(sdkmath.NewInt(1000))
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	expectedErr := fmt.Sprintf("balance is lower than staking amount. staking amount: %v, balance: %v: insufficient funds", balance.Add(sdkmath.NewInt(1000)), balance)
	s.Require().EqualError(err, expectedErr)
}

func (s *KeeperTestSuite) TestLiquidStake_HaltedZone() {
	tc := s.SetupLiquidStake()
	haltedHostZone := tc.initialState.hostZone
	haltedHostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haltedHostZone)
	s.FundAccount(tc.user.acc, sdk.NewInt64Coin(haltedHostZone.IbcDenom, 1000000000))
	_, err := s.GetMsgServer().LiquidStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)

	s.Require().EqualError(err, fmt.Sprintf("halted host zone found for denom (%s): Halted host zone found", haltedHostZone.HostDenom))
}

// ----------------------------------------------------
//	                 RedeemStake
// ----------------------------------------------------

type RedeemStakeState struct {
	epochNumber                        uint64
	initialNativeEpochUnbondingAmount  sdkmath.Int
	initialStTokenEpochUnbondingAmount sdkmath.Int
}
type RedeemStakeTestCase struct {
	user                 Account
	hostZone             stakeibctypes.HostZone
	zoneAccount          Account
	initialState         RedeemStakeState
	validMsg             stakeibctypes.MsgRedeemStake
	expectedNativeAmount sdkmath.Int
}

func (s *KeeperTestSuite) SetupRedeemStake() RedeemStakeTestCase {
	redeemAmount := sdkmath.NewInt(1_000_000)
	redemptionRate := sdk.MustNewDecFromStr("1.5")
	expectedNativeAmount := sdkmath.NewInt(1_500_000)

	user := Account{
		acc:           s.TestAccs[0],
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(user.acc, user.atomBalance)
	s.FundAccount(user.acc, user.stAtomBalance)

	depositAddress := stakeibctypes.NewHostZoneDepositAddress(HostChainId)

	zoneAccount := Account{
		acc:           depositAddress,
		atomBalance:   sdk.NewInt64Coin("ibc/uatom", 10_000_000),
		stAtomBalance: sdk.NewInt64Coin("stuatom", 10_000_000),
	}
	s.FundAccount(zoneAccount.acc, zoneAccount.atomBalance)
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	// TODO define the host zone with total delegation and validators with staked amounts
	hostZone := stakeibctypes.HostZone{
		ChainId:          HostChainId,
		HostDenom:        "uatom",
		Bech32Prefix:     "cosmos",
		RedemptionRate:   redemptionRate,
		TotalDelegations: sdkmath.NewInt(1234567890),
		DepositAddress:   depositAddress.String(),
	}

	epochTrackerDay := stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     1,
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}

	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		Denom:             "uatom",
		HostZoneId:        HostChainId,
		Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTrackerDay)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	return RedeemStakeTestCase{
		user:                 user,
		hostZone:             hostZone,
		zoneAccount:          zoneAccount,
		expectedNativeAmount: expectedNativeAmount,
		initialState: RedeemStakeState{
			epochNumber:                        epochTrackerDay.EpochNumber,
			initialNativeEpochUnbondingAmount:  sdkmath.ZeroInt(),
			initialStTokenEpochUnbondingAmount: sdkmath.ZeroInt(),
		},
		validMsg: stakeibctypes.MsgRedeemStake{
			Creator:  user.acc.String(),
			Amount:   redeemAmount,
			HostZone: HostChainId,
			// TODO set this dynamically through test helpers for host zone
			Receiver: "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf",
		},
	}
}

func (s *KeeperTestSuite) TestRedeemStake_Successful() {
	tc := s.SetupRedeemStake()
	initialState := tc.initialState

	msg := tc.validMsg
	user := tc.user
	redeemAmount := msg.Amount

	// Split the message amount in 2, and call redeem stake twice (each with half the amount)
	// This will check that the same user can redeem multiple times
	msg1 := msg
	msg1.Amount = msg1.Amount.Quo(sdkmath.NewInt(2)) // half the amount

	msg2 := msg
	msg2.Amount = msg.Amount.Sub(msg1.Amount) // remaining half

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &msg1)
	s.Require().NoError(err, "no error expected during first redemption")

	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &msg2)
	s.Require().NoError(err, "no error expected during second redemption")

	// User STUATOM balance should have DECREASED by the amount to be redeemed
	expectedUserStAtomBalance := user.stAtomBalance.SubAmount(redeemAmount)
	actualUserStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, user.acc, "stuatom")
	s.CompareCoins(expectedUserStAtomBalance, actualUserStAtomBalance, "user stuatom balance")

	// Gaia's hostZoneUnbonding NATIVE TOKEN amount should have INCREASED from 0 to the amount redeemed multiplied by the redemption rate
	// Gaia's hostZoneUnbonding STTOKEN amount should have INCREASED from 0 to be amount redeemed
	epochTracker, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx, "day")
	s.Require().True(found, "epoch tracker")
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, epochTracker.EpochNumber)
	s.Require().True(found, "epoch unbonding record")
	hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochUnbondingRecord.EpochNumber, HostChainId)
	s.Require().True(found, "host zone unbondings by chain ID")

	expectedHostZoneUnbondingNativeAmount := initialState.initialNativeEpochUnbondingAmount.Add(tc.expectedNativeAmount)
	expectedHostZoneUnbondingStTokenAmount := initialState.initialStTokenEpochUnbondingAmount.Add(redeemAmount)

	s.Require().Equal(expectedHostZoneUnbondingNativeAmount, hostZoneUnbonding.NativeTokenAmount, "host zone native unbonding amount")
	s.Require().Equal(expectedHostZoneUnbondingStTokenAmount, hostZoneUnbonding.StTokenAmount, "host zone stToken burn amount")

	// UserRedemptionRecord should have been created with correct amount, sender, receiver, host zone, claimIsPending
	userRedemptionRecords := hostZoneUnbonding.UserRedemptionRecords
	s.Require().Equal(len(userRedemptionRecords), 1)
	userRedemptionRecordId := userRedemptionRecords[0]
	userRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, userRedemptionRecordId)
	s.Require().True(found)

	s.Require().Equal(msg.Amount, userRedemptionRecord.StTokenAmount, "redemption record sttoken amount")
	s.Require().Equal(tc.expectedNativeAmount, userRedemptionRecord.NativeTokenAmount, "redemption record native amount")
	s.Require().Equal(msg.Receiver, userRedemptionRecord.Receiver, "redemption record receiver")
	s.Require().Equal(msg.HostZone, userRedemptionRecord.HostZoneId, "redemption record host zone")
	s.Require().False(userRedemptionRecord.ClaimIsPending, "redemption record is not claimable")
	s.Require().NotEqual(hostZoneUnbonding.Status, recordtypes.HostZoneUnbonding_CLAIMABLE, "host zone unbonding should NOT be marked as CLAIMABLE")
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidCreatorAddress() {
	tc := s.SetupRedeemStake()
	invalidMsg := tc.validMsg

	// cosmos instead of stride address
	invalidMsg.Creator = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: invalid Bech32 prefix; expected stride, got cosmos: invalid address", invalidMsg.Creator))

	// invalid stride address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid checksum (expected 8dpmg9 got yxp8uf): invalid address", invalidMsg.Creator))

	// empty address
	invalidMsg.Creator = ""
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: empty address string is not allowed: invalid address", invalidMsg.Creator))

	// wrong len address
	invalidMsg.Creator = "stride1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufabc"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, fmt.Sprintf("creator address is invalid: %s. err: decoding bech32 failed: invalid character not part of charset: 98: invalid address", invalidMsg.Creator))
}

func (s *KeeperTestSuite) TestRedeemStake_HostZoneNotFound() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.HostZone = "fake_host_zone"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone is invalid: fake_host_zone: host zone not registered")
}

func (s *KeeperTestSuite) TestRedeemStake_RateAboveMaxThreshold() {
	tc := s.SetupRedeemStake()

	hz := tc.hostZone
	hz.RedemptionRate = sdk.NewDec(100)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidReceiverAddress() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg

	// stride instead of cosmos address
	invalidMsg.Receiver = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (invalid Bech32 prefix; expected cosmos, got stride): invalid address")

	// invalid cosmos address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ua"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected yxp8uf got yxp8ua)): invalid address")

	// empty address
	invalidMsg.Receiver = ""
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (empty address string is not allowed): invalid address")

	// wrong len address
	invalidMsg.Receiver = "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8ufa"
	_, err = s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().EqualError(err, "invalid receiver address (decoding bech32 failed: invalid checksum (expected xp8ugp got xp8ufa)): invalid address")
}

func (s *KeeperTestSuite) TestRedeemStake_RedeemMoreThanStaked() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	invalidMsg.Amount = sdkmath.NewInt(1_000_000_000_000_000)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, fmt.Sprintf("cannot unstake an amount g.t. staked balance on host zone: %v: invalid amount", invalidMsg.Amount))
}

func (s *KeeperTestSuite) TestRedeemStake_NoEpochTrackerDay() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, tc.initialState.epochNumber)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "latest epoch unbonding record not found: epoch unbonding record not found")
}

func (s *KeeperTestSuite) TestRedeemStake_HostZoneNoUnbondings() {
	tc := s.SetupRedeemStake()

	invalidMsg := tc.validMsg
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
		NativeTokenAmount: sdkmath.ZeroInt(),
		Denom:             "uatom",
		HostZoneId:        "NOT_GAIA",
	}
	epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)

	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &invalidMsg)

	s.Require().EqualError(err, "host zone not found in unbondings: GAIA: host zone not registered")
}

func (s *KeeperTestSuite) TestRedeemStake_InvalidHostAddress() {
	tc := s.SetupRedeemStake()

	// Update hostzone with invalid address
	badHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	badHostZone.DepositAddress = "cosmosXXX"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "could not bech32 decode address cosmosXXX of zone with id: GAIA")
}

func (s *KeeperTestSuite) TestRedeemStake_HaltedZone() {
	tc := s.SetupRedeemStake()

	// Update hostzone with halted
	haltedHostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.validMsg.HostZone)
	haltedHostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haltedHostZone)

	_, err := s.GetMsgServer().RedeemStake(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "halted host zone found for zone (GAIA): Halted host zone found")
}

type LSMLiquidStakeTestCase struct {
	hostZone             types.HostZone
	liquidStakerAddress  sdk.AccAddress
	depositAddress       sdk.AccAddress
	initialBalance       sdkmath.Int
	initialQueryProgress sdkmath.Int
	queryCheckpoint      sdkmath.Int
	lsmTokenIBCDenom     string
	validMsg             *types.MsgLSMLiquidStake
}

// Helper function to add the port and channel onto the LSMTokenBaseDenom,
// hash it, and then store the trace in the IBC store
// Returns the ibc hash
func (s *KeeperTestSuite) getLSMTokenIBCDenom() string {
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, ibctesting.FirstChannelID)
	prefixedDenom := sourcePrefix + LSMTokenBaseDenom
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)
	return lsmTokenDenomTrace.IBCDenom()
}

func (s *KeeperTestSuite) SetupTestLSMLiquidStake() LSMLiquidStakeTestCase {
	initialBalance := sdkmath.NewInt(3000)
	stakeAmount := sdkmath.NewInt(1000)
	userAddress := s.TestAccs[0]
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)

	// Need valid IBC denom here to test parsing
	lsmTokenIBCDenom := s.getLSMTokenIBCDenom()

	// Fund the user's account with the LSM token
	s.FundAccount(userAddress, sdk.NewCoin(lsmTokenIBCDenom, initialBalance))

	// Add the slash interval
	// TVL: 100k, Checkpoint: 1% of 1M = 10k
	// Progress towards query: 8000
	// => Liquid Stake of 2k will trip query
	totalHostZoneStake := sdkmath.NewInt(1_000_000)
	queryCheckpoint := sdkmath.NewInt(10_000)
	progressTowardsQuery := sdkmath.NewInt(8000)
	params := types.DefaultParams()
	params.ValidatorSlashQueryThreshold = 1 // 1 %
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// Sanity check
	onePercent := sdk.MustNewDecFromStr("0.01")
	s.Require().Equal(queryCheckpoint.Int64(), onePercent.Mul(sdk.NewDecFromInt(totalHostZoneStake)).TruncateInt64(),
		"setup failed - query checkpoint must be 1% of total host zone stake")

	// Add the host zone with a valid zone address as the LSM custodian
	hostZone := types.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		RedemptionRate:    sdk.NewDec(1.0),
		DepositAddress:    depositAddress.String(),
		TransferChannelId: ibctesting.FirstChannelID,
		ConnectionId:      ibctesting.FirstConnectionID,
		TotalDelegations:  totalHostZoneStake,
		Validators: []*types.Validator{{
			Address:                   ValAddress,
			SlashQueryProgressTracker: progressTowardsQuery,
			SlashQueryCheckpoint:      queryCheckpoint,
			SharesToTokensRate:        sdk.OneDec(),
		}},
		DelegationIcaAddress:  "cosmos_DELEGATION",
		LsmLiquidStakeEnabled: true,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the latest client height for the ICQ submission
	s.MockClientLatestHeight(1)

	return LSMLiquidStakeTestCase{
		hostZone:             hostZone,
		liquidStakerAddress:  userAddress,
		depositAddress:       depositAddress,
		initialBalance:       initialBalance,
		initialQueryProgress: progressTowardsQuery,
		queryCheckpoint:      queryCheckpoint,
		lsmTokenIBCDenom:     lsmTokenIBCDenom,
		validMsg: &types.MsgLSMLiquidStake{
			Creator:          userAddress.String(),
			LsmTokenIbcDenom: lsmTokenIBCDenom,
			Amount:           stakeAmount,
		},
	}
}

func (s *KeeperTestSuite) TestLSMLiquidStake_Successful_NoSharesToTokensRateQuery() {
	tc := s.SetupTestLSMLiquidStake()

	// Call LSM Liquid stake with a valid message
	msgResponse, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().NoError(err, "no error expected when calling lsm liquid stake")
	s.Require().True(msgResponse.TransactionComplete, "transaction should be complete")

	// Confirm the LSM token was sent to the protocol
	userLsmBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, tc.lsmTokenIBCDenom)
	s.Require().Equal(tc.initialBalance.Sub(tc.validMsg.Amount).Int64(), userLsmBalance.Amount.Int64(),
		"lsm token balance of user account")

	// Confirm stToken was sent to the user
	userStTokenBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom)
	s.Require().Equal(tc.validMsg.Amount.Int64(), userStTokenBalance.Amount.Int64(), "user stToken balance")

	// Confirm an LSMDeposit was created
	expectedDepositId := keeper.GetLSMTokenDepositId(s.Ctx.BlockHeight(), HostChainId, tc.validMsg.Creator, LSMTokenBaseDenom)
	expectedDeposit := recordstypes.LSMTokenDeposit{
		DepositId:        expectedDepositId,
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		StakerAddress:    s.TestAccs[0].String(),
		IbcDenom:         tc.lsmTokenIBCDenom,
		ValidatorAddress: ValAddress,
		Amount:           tc.validMsg.Amount,
		Status:           recordstypes.LSMTokenDeposit_TRANSFER_QUEUE,
		StToken:          sdk.NewCoin(StAtom, tc.validMsg.Amount),
	}
	actualDeposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "lsm token deposit should have been found after LSM liquid stake")
	s.Require().Equal(expectedDeposit, actualDeposit)

	// Confirm slash query progress was incremented
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	expectedQueryProgress := tc.initialQueryProgress.Add(tc.validMsg.Amount)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(expectedQueryProgress.Int64(), hostZone.Validators[0].SlashQueryProgressTracker.Int64(), "slash query progress")
}

func (s *KeeperTestSuite) TestLSMLiquidStake_Successful_WithSharesToTokensRateQuery() {
	tc := s.SetupTestLSMLiquidStake()

	// Increase the liquid stake size so that it breaks the query checkpoint
	// queryProgressSlack is the remaining amount that can be staked in one message before a slash query is issued
	queryProgressSlack := tc.queryCheckpoint.Sub(tc.initialQueryProgress)
	tc.validMsg.Amount = queryProgressSlack.Add(sdk.NewInt(1000))

	// Call LSM Liquid stake
	msgResponse, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().NoError(err, "no error expected when calling lsm liquid stake")
	s.Require().False(msgResponse.TransactionComplete, "transaction should still be pending")

	// Confirm stToken was NOT sent to the user
	userStTokenBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom)
	s.Require().True(userStTokenBalance.Amount.IsZero(), "user stToken balance")

	// Confirm query was submitted
	allQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(allQueries, 1)

	// Confirm query metadata
	actualQuery := allQueries[0]
	s.Require().Equal(HostChainId, actualQuery.ChainId, "query chain-id")
	s.Require().Equal(ibctesting.FirstConnectionID, actualQuery.ConnectionId, "query connection-id")
	s.Require().Equal(icqtypes.STAKING_STORE_QUERY_WITH_PROOF, actualQuery.QueryType, "query types")

	s.Require().Equal(types.ModuleName, actualQuery.CallbackModule, "callback module")
	s.Require().Equal(keeper.ICQCallbackID_Validator, actualQuery.CallbackId, "callback-id")

	expectedTimeout := uint64(s.Ctx.BlockTime().UnixNano() + (keeper.LSMSlashQueryTimeout).Nanoseconds())
	s.Require().Equal(keeper.LSMSlashQueryTimeout, actualQuery.TimeoutDuration, "timeout duration")
	s.Require().Equal(int64(expectedTimeout), int64(actualQuery.TimeoutTimestamp), "timeout timestamp")

	// Confirm query callback data
	s.Require().True(len(actualQuery.CallbackData) > 0, "callback data exists")

	expectedStToken := sdk.NewCoin(StAtom, tc.validMsg.Amount)
	expectedDepositId := keeper.GetLSMTokenDepositId(s.Ctx.BlockHeight(), HostChainId, tc.validMsg.Creator, LSMTokenBaseDenom)
	expectedLSMTokenDeposit := recordstypes.LSMTokenDeposit{
		DepositId:        expectedDepositId,
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		IbcDenom:         tc.lsmTokenIBCDenom,
		StakerAddress:    tc.validMsg.Creator,
		ValidatorAddress: ValAddress,
		Amount:           tc.validMsg.Amount,
		StToken:          expectedStToken,
		Status:           recordstypes.LSMTokenDeposit_DEPOSIT_PENDING,
	}

	var actualCallbackData types.ValidatorSharesToTokensQueryCallback
	err = proto.Unmarshal(actualQuery.CallbackData, &actualCallbackData)
	s.Require().NoError(err, "no error expected when unmarshalling query callback data")

	lsmLiquidStake := actualCallbackData.LsmLiquidStake
	s.Require().Equal(HostChainId, lsmLiquidStake.HostZone.ChainId, "callback data - host zone")
	s.Require().Equal(ValAddress, lsmLiquidStake.Validator.Address, "callback data - validator")

	s.Require().Equal(expectedLSMTokenDeposit, *lsmLiquidStake.Deposit, "callback data - deposit")
}

func (s *KeeperTestSuite) TestLSMLiquidStake_DifferentRedemptionRates() {
	tc := s.SetupTestLSMLiquidStake()
	tc.validMsg.Amount = sdk.NewInt(100) // reduce the stake amount to prevent insufficient balance error

	// Loop over sharesToTokens rates: {0.92, 0.94, ..., 1.2}
	interval := sdk.MustNewDecFromStr("0.01")
	for i := -8; i <= 10; i += 2 {
		redemptionDelta := interval.Mul(sdk.NewDec(int64(i))) // i = 2 => delta = 0.02
		newRedemptionRate := sdk.NewDec(1.0).Add(redemptionDelta)
		redemptionRateFloat := newRedemptionRate

		// Update rate in host zone
		hz := tc.hostZone
		hz.RedemptionRate = newRedemptionRate
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hz)

		// Liquid stake for each balance and confirm stAtom minted
		startingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom).Amount
		_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
		s.Require().NoError(err)
		endingStAtomBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.liquidStakerAddress, StAtom).Amount
		actualStAtomMinted := endingStAtomBalance.Sub(startingStAtomBalance)

		expectedStAtomMinted := sdk.NewDecFromInt(tc.validMsg.Amount).Quo(redemptionRateFloat).TruncateInt()
		testDescription := fmt.Sprintf("st atom balance for redemption rate: %v", redemptionRateFloat)
		s.Require().Equal(expectedStAtomMinted, actualStAtomMinted, testDescription)

		// Cleanup the LSMTokenDeposit record to prevent an error on the next run
		s.App.RecordsKeeper.RemoveLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	}
}

// ----------------------------------------------------
//	               PrepareDelegation
// ----------------------------------------------------

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_NotIBCDenom() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the denom is not an IBC token
	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = "fake_ibc_denom"

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "lsm token is not an IBC token (fake_ibc_denom)")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_HostZoneNotFound() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the denom is an IBC denom from a channel that is not supported
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, "channel-1")
	prefixedDenom := sourcePrefix + LSMTokenBaseDenom
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)

	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = lsmTokenDenomTrace.IBCDenom()

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-1) does not match any registered host zone")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_ValidatorNotFound() {
	tc := s.SetupTestLSMLiquidStake()

	// Change the message so that the base denom is from a non-existent validator
	sourcePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, ibctesting.FirstChannelID)
	prefixedDenom := sourcePrefix + "cosmosvaloperXXX/42"
	lsmTokenDenomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, lsmTokenDenomTrace)

	invalidMsg := tc.validMsg
	invalidMsg.LsmTokenIbcDenom = lsmTokenDenomTrace.IBCDenom()

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), invalidMsg)
	s.Require().ErrorContains(err, "validator (cosmosvaloperXXX) is not registered in the Stride validator set")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_DepositAlreadyExists() {
	tc := s.SetupTestLSMLiquidStake()

	// Set a deposit with the same chainID and denom in the store
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordstypes.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   LSMTokenBaseDenom,
	})

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "there is already a previous record with this denom being processed")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_InvalidDepositAddress() {
	tc := s.SetupTestLSMLiquidStake()

	// Remove the host zones address from the store
	invalidHostZone := tc.hostZone
	invalidHostZone.DepositAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "host zone address is invalid")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_InsufficientBalance() {
	tc := s.SetupTestLSMLiquidStake()

	// Send out all the user's coins so that they have an insufficient balance of LSM tokens
	initialBalanceCoin := sdk.NewCoins(sdk.NewCoin(tc.lsmTokenIBCDenom, tc.initialBalance))
	err := s.App.BankKeeper.SendCoins(s.Ctx, tc.liquidStakerAddress, s.TestAccs[1], initialBalanceCoin)
	s.Require().NoError(err)

	_, err = s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().ErrorContains(err, "insufficient funds")
}

func (s *KeeperTestSuite) TestLSMLiquidStakeFailed_ZeroStTokens() {
	tc := s.SetupTestLSMLiquidStake()

	// Adjust redemption rate and liquid stake amount so that the number of stTokens would be zero
	// stTokens = 1(amount) / 1.1(RR) = rounds down to 0
	hostZone := tc.hostZone
	hostZone.RedemptionRate = sdk.NewDecWithPrec(11, 1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	tc.validMsg.Amount = sdkmath.NewInt(1)

	// The liquid stake should fail
	_, err := s.GetMsgServer().LSMLiquidStake(sdk.WrapSDKContext(s.Ctx), tc.validMsg)
	s.Require().EqualError(err, "Liquid stake of 1uatom would return 0 stTokens: Liquid staked amount is too small")
}

// ----------------------------------------------------
//	               DeleteTradeRoute
// ----------------------------------------------------

func (s *KeeperTestSuite) TestDeleteTradeRoute() {
	initialRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, initialRoute)

	msg := types.MsgDeleteTradeRoute{
		Authority:   Authority,
		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,
	}

	// Confirm the route is present before attepmting to delete was deleted
	_, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been found before delete message")

	// Delete the trade route
	_, err := s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when deleting trade route")

	// Confirm it was deleted
	_, found = s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().False(found, "trade route should have been deleted")

	// Attempt to delete it again, it should fail since it doesn't exist
	_, err = s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "trade route not found")

	// Attempt to delete with the wrong authority - it should fail
	invalidMsg := msg
	invalidMsg.Authority = "not-gov-address"

	_, err = s.GetMsgServer().DeleteTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid authority")
}

// ----------------------------------------------------
//	               CreateTradeRoute
// ----------------------------------------------------

func (s *KeeperTestSuite) SetupTestCreateTradeRoute() (msg types.MsgCreateTradeRoute, expectedTradeRoute types.TradeRoute) {
	rewardChainId := "reward-0"
	tradeChainId := "trade-0"

	hostConnectionId := "connection-0"
	rewardConnectionId := "connection-1"
	tradeConnectionId := "connection-2"

	hostToRewardChannelId := "channel-100"
	rewardToTradeChannelId := "channel-200"
	tradeToHostChannelId := "channel-300"

	rewardDenomOnHost := "ibc/reward-on-host"
	rewardDenomOnReward := RewardDenom
	rewardDenomOnTrade := "ibc/reward-on-trade"
	hostDenomOnTrade := "ibc/host-on-trade"
	hostDenomOnHost := HostDenom

	withdrawalAddress := "withdrawal-address"
	unwindAddress := "unwind-address"

	minTransferAmount := sdkmath.NewInt(100)

	// Register an exisiting ICA account for the unwind ICA to test that
	// existing accounts are re-used
	owner := types.FormatTradeRouteICAOwner(rewardChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_UNWIND)
	s.MockICAChannel(rewardConnectionId, "channel-0", owner, unwindAddress)

	// Mock out connections for the reward an trade chain so that an ICA registration can be submitted
	s.MockClientAndConnection(rewardChainId, "07-tendermint-0", rewardConnectionId)
	s.MockClientAndConnection(tradeChainId, "07-tendermint-1", tradeConnectionId)

	// Create a host zone with an exisiting withdrawal address
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         hostConnectionId,
		WithdrawalIcaAddress: withdrawalAddress,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Define a valid message given the parameters above
	msg = types.MsgCreateTradeRoute{
		Authority:   Authority,
		HostChainId: HostChainId,

		StrideToRewardConnectionId: rewardConnectionId,
		StrideToTradeConnectionId:  tradeConnectionId,

		HostToRewardTransferChannelId:  hostToRewardChannelId,
		RewardToTradeTransferChannelId: rewardToTradeChannelId,
		TradeToHostTransferChannelId:   tradeToHostChannelId,

		RewardDenomOnHost:   rewardDenomOnHost,
		RewardDenomOnReward: rewardDenomOnReward,
		RewardDenomOnTrade:  rewardDenomOnTrade,
		HostDenomOnTrade:    hostDenomOnTrade,
		HostDenomOnHost:     hostDenomOnHost,

		MinTransferAmount: minTransferAmount,
	}

	// Build out the expected trade route given the above
	expectedTradeRoute = types.TradeRoute{
		RewardDenomOnHostZone:   rewardDenomOnHost,
		RewardDenomOnRewardZone: rewardDenomOnReward,
		RewardDenomOnTradeZone:  rewardDenomOnTrade,
		HostDenomOnTradeZone:    hostDenomOnTrade,
		HostDenomOnHostZone:     hostDenomOnHost,

		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Type:         types.ICAAccountType_WITHDRAWAL,
			ConnectionId: hostConnectionId,
			Address:      withdrawalAddress,
		},
		RewardAccount: types.ICAAccount{
			ChainId:      rewardChainId,
			Type:         types.ICAAccountType_CONVERTER_UNWIND,
			ConnectionId: rewardConnectionId,
			Address:      unwindAddress,
		},
		TradeAccount: types.ICAAccount{
			ChainId:      tradeChainId,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
			ConnectionId: tradeConnectionId,
		},

		HostToRewardChannelId:  hostToRewardChannelId,
		RewardToTradeChannelId: rewardToTradeChannelId,
		TradeToHostChannelId:   tradeToHostChannelId,

		MinTransferAmount: minTransferAmount,
	}

	return msg, expectedTradeRoute
}

// Helper function to create a trade route and check the created route matched expectations
func (s *KeeperTestSuite) submitCreateTradeRouteAndValidate(msg types.MsgCreateTradeRoute, expectedRoute types.TradeRoute) {
	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when creating trade route")

	actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, msg.RewardDenomOnReward, msg.HostDenomOnHost)
	s.Require().True(found, "trade route should have been created")

	s.Require().Equal(expectedRoute.RewardDenomOnHostZone, actualRoute.RewardDenomOnHostZone, "trade route reward on host denom")
	s.Require().Equal(expectedRoute.RewardDenomOnRewardZone, actualRoute.RewardDenomOnRewardZone, "trade route reward on reward denom")
	s.Require().Equal(expectedRoute.RewardDenomOnTradeZone, actualRoute.RewardDenomOnTradeZone, "trade route reward on trade denom")
	s.Require().Equal(expectedRoute.HostDenomOnTradeZone, actualRoute.HostDenomOnTradeZone, "trade route host on trade denom")
	s.Require().Equal(expectedRoute.HostDenomOnHostZone, actualRoute.HostDenomOnHostZone, "trade route host on host denom")

	s.Require().Equal(expectedRoute.HostAccount, actualRoute.HostAccount, "trade route host account")
	s.Require().Equal(expectedRoute.RewardAccount, actualRoute.RewardAccount, "trade route reward account")
	s.Require().Equal(expectedRoute.TradeAccount, actualRoute.TradeAccount, "trade route trade account")

	s.Require().Equal(expectedRoute.HostToRewardChannelId, actualRoute.HostToRewardChannelId, "trade route host to reward")
	s.Require().Equal(expectedRoute.RewardToTradeChannelId, actualRoute.RewardToTradeChannelId, "trade route reward to trade")
	s.Require().Equal(expectedRoute.TradeToHostChannelId, actualRoute.TradeToHostChannelId, "trade route trade to host")

	s.Require().Equal(expectedRoute.MinTransferAmount, actualRoute.MinTransferAmount, "trade route min transfer amount")
}

// Tests a successful trade route creation
func (s *KeeperTestSuite) TestCreateTradeRoute_Success() {
	msg, expectedRoute := s.SetupTestCreateTradeRoute()
	s.submitCreateTradeRouteAndValidate(msg, expectedRoute)
}

// Tests trying to create a route from an invalid authority
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_Authority() {
	msg, _ := s.SetupTestCreateTradeRoute()

	msg.Authority = "not-gov-address"

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "invalid authority")
}

// Tests creating a duplicate trade route
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_DuplicateTradeRoute() {
	msg, _ := s.SetupTestCreateTradeRoute()

	// Store down a trade route so the tx hits a duplicate trade route error
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	})

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "Trade route already exists")
}

// Tests creating a trade route when the host zone or withdrawal address does not exist
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_HostZoneNotRegistered() {
	msg, _ := s.SetupTestCreateTradeRoute()

	// Remove the host zone withdrawal address and confirm it fails
	invalidHostZone := s.MustGetHostZone(HostChainId)
	invalidHostZone.WithdrawalIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "withdrawal account not initialized on host zone")

	// Remove the host zone completely and check that that also fails
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	_, err = s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "host zone not found")
}

// Tests creating a trade route where the ICA channels cannot be created
// because the ICA connections do not exist
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_ConnectionNotFound() {
	// Test with non-existent reward connection
	msg, _ := s.SetupTestCreateTradeRoute()
	msg.StrideToRewardConnectionId = "connection-X"

	// Remove the host zone completely and check that that also fails
	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the unwind ICA account: connection connection-X not found")

	// Setup again, but this time use a non-existent trade connection
	msg, _ = s.SetupTestCreateTradeRoute()
	msg.StrideToTradeConnectionId = "connection-Y"

	_, err = s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the trade ICA account: connection connection-Y not found")
}

// Tests creating a trade route where the ICA registration step fails
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_UnableToRegisterICA() {
	msg, expectedRoute := s.SetupTestCreateTradeRoute()

	// Disable ICA middleware for the trade channel so the ICA fails
	tradeAccount := expectedRoute.TradeAccount
	tradeOwner := types.FormatTradeRouteICAOwner(tradeAccount.ChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	tradePortId, _ := icatypes.NewControllerPortID(tradeOwner)
	s.App.ICAControllerKeeper.SetMiddlewareDisabled(s.Ctx, tradePortId, tradeAccount.ConnectionId)

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the trade ICA account")
}

// ----------------------------------------------------
//	               UpdateTradeRoute
// ----------------------------------------------------

// Helper function to update a trade route and check the updated route matched expectations
func (s *KeeperTestSuite) submitUpdateTradeRouteAndValidate(msg types.MsgUpdateTradeRoute, expectedRoute types.TradeRoute) {
	_, err := s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when updating trade route")

	actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been updated")
	s.Require().Equal(expectedRoute.RewardDenomOnRewardZone, actualRoute.RewardDenomOnRewardZone, "trade route reward denom")
	s.Require().Equal(expectedRoute.HostDenomOnHostZone, actualRoute.HostDenomOnHostZone, "trade route host denom")
	s.Require().Equal(expectedRoute.MinTransferAmount, actualRoute.MinTransferAmount, "trade route min transfer amount")
}

func (s *KeeperTestSuite) TestUpdateTradeRoute() {
	minTransferAmount := sdkmath.NewInt(100)

	// Create a trade route with no parameters
	initialRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, initialRoute)

	// Define a valid message given the parameters above
	msg := types.MsgUpdateTradeRoute{
		Authority:         Authority,
		RewardDenom:       RewardDenom,
		HostDenom:         HostDenom,
		MinTransferAmount: minTransferAmount,
	}

	// Build out the expected trade route given the above
	expectedRoute := initialRoute
	expectedRoute.MinTransferAmount = minTransferAmount

	// Update the route and confirm the changes persisted
	s.submitUpdateTradeRouteAndValidate(msg, expectedRoute)

	// Test that an error is thrown if the correct authority is not specified
	invalidMsg := msg
	invalidMsg.Authority = "not-gov-address"

	_, err := s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid authority")

	// Test that an error is thrown if the route doesn't exist
	invalidMsg = msg
	invalidMsg.RewardDenom = "invalid-reward-denom"

	_, err = s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "trade route not found")
}

// ----------------------------------------------------
//	           RestoreInterchainAccount
// ----------------------------------------------------

type DepositRecordStatusUpdate struct {
	chainId        string
	initialStatus  recordtypes.DepositRecord_Status
	revertedStatus recordtypes.DepositRecord_Status
}

type HostZoneUnbondingStatusUpdate struct {
	initialStatus  recordtypes.HostZoneUnbonding_Status
	revertedStatus recordtypes.HostZoneUnbonding_Status
}

type LSMTokenDepositStatusUpdate struct {
	chainId        string
	denom          string
	initialStatus  recordtypes.LSMTokenDeposit_Status
	revertedStatus recordtypes.LSMTokenDeposit_Status
}

type RestoreInterchainAccountTestCase struct {
	validMsg                    types.MsgRestoreInterchainAccount
	depositRecordStatusUpdates  []DepositRecordStatusUpdate
	unbondingRecordStatusUpdate []HostZoneUnbondingStatusUpdate
	lsmTokenDepositStatusUpdate []LSMTokenDepositStatusUpdate
	delegationChannelID         string
	delegationPortID            string
}

func (s *KeeperTestSuite) SetupRestoreInterchainAccount(createDelegationICAChannel bool) RestoreInterchainAccountTestCase {
	s.CreateTransferChannel(HostChainId)

	// We have to setup the ICA channel before the LSM Token is stored,
	// otherwise when the EndBlocker runs in the channel setup, the LSM Token
	// statuses will get updated
	var channelID, portID string
	if createDelegationICAChannel {
		owner := "GAIA.DELEGATION"
		channelID, portID = s.CreateICAChannel(owner)
	}

	hostZone := types.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		RedemptionRate: sdk.OneDec(), // if not set, the beginblocker invariant panics
		Validators: []*types.Validator{
			{Address: "valA", DelegationChangesInProgress: 1},
			{Address: "valB", DelegationChangesInProgress: 2},
			{Address: "valC", DelegationChangesInProgress: 3},
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store deposit records with some in state pending
	depositRecords := []DepositRecordStatusUpdate{
		{
			// Status doesn't change
			chainId:        HostChainId,
			initialStatus:  recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			chainId:        HostChainId,
			initialStatus:  recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_DELEGATION_QUEUE,
		},
		{
			// Status doesn't get reveted because it's a different host zone
			chainId:        "different_host_zone",
			initialStatus:  recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus: recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
		},
	}
	for i, depositRecord := range depositRecords {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordtypes.DepositRecord{
			Id:         uint64(i),
			HostZoneId: depositRecord.chainId,
			Status:     depositRecord.initialStatus,
		})
	}

	// Store epoch unbonding records with some in state pending
	hostZoneUnbondingRecords := []HostZoneUnbondingStatusUpdate{
		{
			// Status doesn't change
			initialStatus:  recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
			revertedStatus: recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			initialStatus:  recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
			revertedStatus: recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
		},
		{
			// Status doesn't change
			initialStatus:  recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
			revertedStatus: recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			initialStatus:  recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		},
	}
	for i, hostZoneUnbonding := range hostZoneUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: uint64(i),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				// The first unbonding record will get reverted, the other one will not
				{
					HostZoneId:                HostChainId,
					Status:                    hostZoneUnbonding.initialStatus,
					UndelegationTxsInProgress: 4,
				},
				{
					HostZoneId:                "different_host_zone",
					Status:                    hostZoneUnbonding.initialStatus,
					UndelegationTxsInProgress: 5,
				},
			},
		})
	}

	// Store LSM Token Deposits with some state pending
	lsmTokenDeposits := []LSMTokenDepositStatusUpdate{
		{
			// Status doesn't change
			chainId:        HostChainId,
			denom:          "denom-1",
			initialStatus:  recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			chainId:        HostChainId,
			denom:          "denom-2",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		},
		{
			// Status doesn't change
			chainId:        HostChainId,
			denom:          "denom-3",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		},
		{
			// Status doesn't change (different host zone)
			chainId:        "different_host_zone",
			denom:          "denom-4",
			initialStatus:  recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
			revertedStatus: recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
		},
	}
	for _, lsmTokenDeposit := range lsmTokenDeposits {
		s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordtypes.LSMTokenDeposit{
			ChainId: lsmTokenDeposit.chainId,
			Status:  lsmTokenDeposit.initialStatus,
			Denom:   lsmTokenDeposit.denom,
		})
	}

	defaultMsg := types.MsgRestoreInterchainAccount{
		Creator:      "creatoraddress",
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		AccountOwner: types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION),
	}

	return RestoreInterchainAccountTestCase{
		validMsg:                    defaultMsg,
		depositRecordStatusUpdates:  depositRecords,
		unbondingRecordStatusUpdate: hostZoneUnbondingRecords,
		lsmTokenDepositStatusUpdate: lsmTokenDeposits,
		delegationChannelID:         channelID,
		delegationPortID:            portID,
	}
}

// Helper function to close an ICA channel
func (s *KeeperTestSuite) closeICAChannel(portId, channelID string) {
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portId, channelID)
	s.Require().True(found, "unable to close channel because channel was not found")
	channel.State = channeltypes.CLOSED
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, channelID, channel)
}

// Helper function to call RestoreChannel and check that a new channel was created and opened
func (s *KeeperTestSuite) restoreChannelAndVerifySuccess(msg types.MsgRestoreInterchainAccount, portID string, channelID string) {
	// Restore the channel
	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "registered ica account successfully")

	// Confirm channel was created
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
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

// Helper function to check that each DepositRecord's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyDepositRecordsStatus(expectedDepositRecords []DepositRecordStatusUpdate, revert bool) {
	for i, expectedDepositRecord := range expectedDepositRecords {
		actualDepositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, uint64(i))
		s.Require().True(found, "deposit record found")

		// Only revert records if the revert option is passed and the host zone matches
		expectedStatus := expectedDepositRecord.initialStatus
		if revert && actualDepositRecord.HostZoneId == HostChainId {
			expectedStatus = expectedDepositRecord.revertedStatus
		}
		s.Require().Equal(expectedStatus.String(), actualDepositRecord.Status.String(), "deposit record %d status", i)
	}
}

// Helper function to check that each HostZoneUnbonding's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyHostZoneUnbondingStatus(expectedUnbondingRecords []HostZoneUnbondingStatusUpdate, revert bool) {
	for i, expectedUnbonding := range expectedUnbondingRecords {
		epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(i))
		s.Require().True(found, "epoch unbonding record found")

		for _, actualUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
			// Only revert records if the revert option is passed and the host zone matches
			expectedStatus := expectedUnbonding.initialStatus
			if revert && actualUnbonding.HostZoneId == HostChainId {
				expectedStatus = expectedUnbonding.revertedStatus
			}
			s.Require().Equal(expectedStatus.String(), actualUnbonding.Status.String(), "host zone unbonding for epoch %d record status", i)
		}
	}
}

// Helper function to check that each LSMTokenDepoit's status was either left alone or reverted to it's prior status
func (s *KeeperTestSuite) verifyLSMDepositStatus(expectedLSMDeposits []LSMTokenDepositStatusUpdate, revert bool) {
	for i, expectedLSMDeposit := range expectedLSMDeposits {
		actualLSMDeposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, expectedLSMDeposit.chainId, expectedLSMDeposit.denom)
		s.Require().True(found, "lsm deposit found")

		// Only revert record if the revert option is passed and the host zone matches
		expectedStatus := expectedLSMDeposit.initialStatus
		if revert && actualLSMDeposit.ChainId == HostChainId {
			expectedStatus = expectedLSMDeposit.revertedStatus
		}
		s.Require().Equal(expectedStatus.String(), actualLSMDeposit.Status.String(), "lsm deposit %d", i)
	}
}

// Helper function to check that the delegation changes in progress field was reset to 0 for each validator
func (s *KeeperTestSuite) verifyDelegationChangeInProgressReset() {
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Len(hostZone.Validators, 3, "there should be 3 validators on this host zone")

	for _, validator := range hostZone.Validators {
		s.Require().Zero(validator.DelegationChangesInProgress,
			"delegation change in progress should have been reset for validator %s", validator.Address)
	}
}

// Helper function to check that the undelegation changes in progress field was reset to 0
// for each host zone unbonding record
func (s *KeeperTestSuite) verifyUndelegationChangeInProgressReset() {
	for _, epochUnbondingRecord := range s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx) {
		for _, hostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			if hostZoneUnbondingRecord.HostZoneId == HostChainId {
				s.Require().Zero(hostZoneUnbondingRecord.UndelegationTxsInProgress,
					"undelegation changes should have been reset for epoch %d", epochUnbondingRecord.EpochNumber)
			} else {
				s.Require().NotZero(hostZoneUnbondingRecord.UndelegationTxsInProgress,
					"undelegation changes should not have been reset for epoch %d", epochUnbondingRecord.EpochNumber)
			}
		}
	}
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_Success() {
	tc := s.SetupRestoreInterchainAccount(true)

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + delegate)")

	// Close the delegation channel
	s.closeICAChannel(tc.delegationPortID, tc.delegationChannelID)

	// Confirm the new channel was created
	s.restoreChannelAndVerifySuccess(tc.validMsg, tc.delegationPortID, tc.delegationChannelID)

	// Verify the record status' were reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, true)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, true)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, true)
	s.verifyDelegationChangeInProgressReset()
	s.verifyUndelegationChangeInProgressReset()
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_InvalidConnectionId() {
	tc := s.SetupRestoreInterchainAccount(false)

	// Update the connectionId on the host zone so that it doesn't exist
	invalidMsg := tc.validMsg
	invalidMsg.ConnectionId = "fake_connection"

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "connection fake_connection not found")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount(false)

	// Attempt to restore an account that does not exist
	msg := tc.validMsg
	msg.AccountOwner = types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "ICA controller account address not found: GAIA.WITHDRAWAL")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_HostZoneNotFound() {
	tc := s.SetupRestoreInterchainAccount(true)
	s.closeICAChannel(tc.delegationPortID, tc.delegationChannelID)

	// Delete the host zone so the lookup fails
	// (this check only runs for the delegation channel)
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "delegation ICA supplied, but no associated host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_RevertDepositRecords_Failure() {
	tc := s.SetupRestoreInterchainAccount(true)

	_, err := s.GetMsgServer().RestoreInterchainAccount(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "existing active channel channel-1 for portID icacontroller-GAIA.DELEGATION")

	// Verify the record status' were NOT reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, false)
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_NoRecordChange_Success() {
	// Here, we're closing and restoring the withdrawal channel so records should not be reverted
	tc := s.SetupRestoreInterchainAccount(false)
	owner := "GAIA.WITHDRAWAL"
	channelID, portID := s.CreateICAChannel(owner)

	// Confirm there are two channels originally
	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
	s.Require().Len(channels, 2, "there should be 2 channels initially (transfer + withdrawal)")

	// Close the withdrawal channel
	s.closeICAChannel(portID, channelID)

	// Restore the channel
	msg := tc.validMsg
	msg.AccountOwner = types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)
	s.restoreChannelAndVerifySuccess(msg, portID, channelID)

	// Verify the record status' were NOT reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, false)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, false)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, false)
}

// ----------------------------------------------------
//	         UpdateInnerRedemptionRateBounds
// ----------------------------------------------------

type UpdateInnerRedemptionRateBoundsTestCase struct {
	validMsg stakeibctypes.MsgUpdateInnerRedemptionRateBounds
	zone     stakeibctypes.HostZone
}

func (s *KeeperTestSuite) SetupUpdateInnerRedemptionRateBounds() UpdateInnerRedemptionRateBoundsTestCase {
	// Register a host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
		MaxRedemptionRate: sdk.NewDec(15).Quo(sdk.NewDec(10)),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	defaultMsg := stakeibctypes.MsgUpdateInnerRedemptionRateBounds{
		// TODO: does this need to be the admin address?
		Creator:                s.TestAccs[0].String(),
		ChainId:                HostChainId,
		MinInnerRedemptionRate: sdk.NewDec(1),
		MaxInnerRedemptionRate: sdk.NewDec(11).Quo(sdk.NewDec(10)),
	}

	return UpdateInnerRedemptionRateBoundsTestCase{
		validMsg: defaultMsg,
		zone:     hostZone,
	}
}

// Verify that bounds can be set successfully
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_Success() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")
	s.Require().Equal(tc.validMsg.MinInnerRedemptionRate, zone.MinInnerRedemptionRate, "min inner redemption rate should be set")
	s.Require().Equal(tc.validMsg.MaxInnerRedemptionRate, zone.MaxInnerRedemptionRate, "max inner redemption rate should be set")
}

// Setting inner bounds outside of outer bounds should throw an error
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_OutOfBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the min inner bound to be less than the min outer bound
	tc.validMsg.MinInnerRedemptionRate = sdk.NewDec(0)

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg := fmt.Sprintf("inner min safety threshold (%s) is less than outer min safety threshold (%s)", tc.validMsg.MinInnerRedemptionRate, sdk.NewDec(9).Quo(sdk.NewDec(10)))
	s.Require().ErrorContains(err, errMsg)

	// Set the min inner bound to be valid, but the max inner bound to be greater than the max outer bound
	tc.validMsg.MinInnerRedemptionRate = sdk.NewDec(1)
	tc.validMsg.MaxInnerRedemptionRate = sdk.NewDec(3)
	// Set the inner bounds on the host zone
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg = fmt.Sprintf("inner max safety threshold (%s) is greater than outer max safety threshold (%s)", tc.validMsg.MaxInnerRedemptionRate, sdk.NewDec(15).Quo(sdk.NewDec(10)))
	s.Require().ErrorContains(err, errMsg)
}

// Validate basic tests
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_InvalidMsg() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the min inner bound to be greater than than the max inner bound
	invalidMsg := tc.validMsg
	invalidMsg.MinInnerRedemptionRate = sdk.NewDec(2)

	err := invalidMsg.ValidateBasic()

	// Verify the error
	errMsg := fmt.Sprintf("Inner max safety threshold (%s) is less than inner min safety threshold (%s)", invalidMsg.MaxInnerRedemptionRate, invalidMsg.MinInnerRedemptionRate)
	s.Require().ErrorContains(err, errMsg)
}

// Verify that if inner bounds end up outside of outer bounds (somehow), the outer bounds are returned
func (s *KeeperTestSuite) TestGetInnerSafetyBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds outside the outer bounds on the host zone directly
	tc.zone.MinInnerRedemptionRate = sdk.NewDec(0)
	tc.zone.MaxInnerRedemptionRate = sdk.NewDec(3)
	// Set the host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.zone)

	// Get the inner bounds and verify the outer bounds are used
	innerMinSafetyThreshold, innerMaxSafetyThreshold := s.App.StakeibcKeeper.GetInnerSafetyBounds(s.Ctx, tc.zone)
	s.Require().Equal(tc.zone.MinRedemptionRate, innerMinSafetyThreshold, "min inner redemption rate should be set")
	s.Require().Equal(tc.zone.MaxRedemptionRate, innerMaxSafetyThreshold, "max inner redemption rate should be set")
}

// ----------------------------------------------------
//	                 ResumeHostZone
// ----------------------------------------------------

type ResumeHostZoneTestCase struct {
	validMsg stakeibctypes.MsgResumeHostZone
	zone     stakeibctypes.HostZone
}

func (s *KeeperTestSuite) SetupResumeHostZone() ResumeHostZoneTestCase {
	// Register a host zone
	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		MinRedemptionRate: sdk.NewDec(9).Quo(sdk.NewDec(10)),
		MaxRedemptionRate: sdk.NewDec(15).Quo(sdk.NewDec(10)),
		Halted:            true,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	defaultMsg := stakeibctypes.MsgResumeHostZone{
		Creator: s.TestAccs[0].String(),
		ChainId: HostChainId,
	}

	return ResumeHostZoneTestCase{
		validMsg: defaultMsg,
		zone:     hostZone,
	}
}

// Verify that bounds can be set successfully
func (s *KeeperTestSuite) TestResumeHostZone_Success() {
	tc := s.SetupResumeHostZone()

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &tc.validMsg)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the inner bounds were set
	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")

	s.Require().False(zone.Halted, "host zone should not be halted")
}

// verify that non-admins can't call the tx
func (s *KeeperTestSuite) TestResumeHostZone_NonAdmin() {
	tc := s.SetupResumeHostZone()

	invalidMsg := tc.validMsg
	invalidMsg.Creator = s.TestAccs[1].String()

	err := invalidMsg.ValidateBasic()
	s.Require().Error(err, "nonadmins shouldn't be able to call this tx")
}

// verify that the function can't be called on missing zones
func (s *KeeperTestSuite) TestResumeHostZone_MissingZones() {
	tc := s.SetupResumeHostZone()

	invalidMsg := tc.validMsg
	invalidChainId := "invalid-chain"
	invalidMsg.ChainId = invalidChainId

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &invalidMsg)

	s.Require().Error(err, "shouldn't be able to call tx on missing zones")
	expectedErrorMsg := fmt.Sprintf("invalid chain id, zone for %s not found: host zone not found", invalidChainId)
	s.Require().Equal(expectedErrorMsg, err.Error(), "should return correct error msg")
}

// verify that the function can't be called on unhalted zones
func (s *KeeperTestSuite) TestResumeHostZone_UnhaltedZones() {
	tc := s.SetupResumeHostZone()

	zone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should be in the store")
	s.Require().True(zone.Halted, "host zone should be halted")
	zone.Halted = false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, zone)

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().ResumeHostZone(s.Ctx, &tc.validMsg)
	s.Require().Error(err, "shouldn't be able to call tx on unhalted zones")
	expectedErrorMsg := fmt.Sprintf("invalid chain id, zone for %s not halted: host zone is not halted", HostChainId)
	s.Require().Equal(expectedErrorMsg, err.Error(), "should return correct error msg")
}

// ----------------------------------------------------
//	           SetCommunityPoolRebate
// ----------------------------------------------------

func (s *KeeperTestSuite) TestSetCommunityPoolRebate() {
	stTokenSupply := sdk.NewInt(2000)
	rebateInfo := types.CommunityPoolRebate{
		RebateRate:                sdk.MustNewDecFromStr("0.5"),
		LiquidStakedStTokenAmount: sdk.NewInt(1000),
	}

	// Mint stTokens so the supply is populated
	s.FundAccount(s.TestAccs[0], sdk.NewCoin(utils.StAssetDenomFromHostZoneDenom(HostDenom), stTokenSupply))

	// Set host zone with no rebate
	hostZone := types.HostZone{
		ChainId:   HostChainId,
		HostDenom: HostDenom,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Submit a message to create the rebate
	registerMsg := types.MsgSetCommunityPoolRebate{
		ChainId:                   HostChainId,
		RebateRate:                rebateInfo.RebateRate,
		LiquidStakedStTokenAmount: rebateInfo.LiquidStakedStTokenAmount,
	}
	_, err := s.GetMsgServer().SetCommunityPoolRebate(s.Ctx, &registerMsg)
	s.Require().NoError(err, "no error expected when registering rebate")

	// Confirm the rebate was updated
	actualHostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(rebateInfo, *actualHostZone.CommunityPoolRebate, "rebate was updated on host zone")

	// Attempt to update the rebate with a large liquid stake amount, it should fail
	invalidMsg := types.MsgSetCommunityPoolRebate{
		ChainId:                   HostChainId,
		LiquidStakedStTokenAmount: sdk.NewInt(1_000_000),
	}
	_, err = s.GetMsgServer().SetCommunityPoolRebate(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "liquid staked stToken amount (1000000) is greater than current supply (2000)")

	// Submit a 0 LS amount which should delete the rebate
	removeMsg := types.MsgSetCommunityPoolRebate{
		ChainId:                   HostChainId,
		LiquidStakedStTokenAmount: sdk.ZeroInt(),
	}
	_, err = s.GetMsgServer().SetCommunityPoolRebate(s.Ctx, &removeMsg)
	s.Require().NoError(err, "no error expected when registering 0 rebate")

	actualHostZone = s.MustGetHostZone(HostChainId)
	s.Require().Nil(actualHostZone.CommunityPoolRebate, "rebate was removed from host zone")

	// Confirm a message with an invalid chain ID would cause an error
	_, err = s.GetMsgServer().SetCommunityPoolRebate(s.Ctx, &types.MsgSetCommunityPoolRebate{ChainId: "invalid"})
	s.Require().ErrorContains(err, "host zone not found")
}

// ----------------------------------------------------
//	              ToggleTradeController
// ----------------------------------------------------

func (s *KeeperTestSuite) TestToggleTradeController() {
	tradeICAOwner := types.FormatTradeRouteICAOwner(HostChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	channelId, portId := s.CreateICAChannel(tradeICAOwner)

	tradeControllerAddress := "trade-controller"

	// Create a trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
		TradeAccount: types.ICAAccount{
			ChainId:      HostChainId,
			ConnectionId: ibctesting.FirstConnectionID,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)

	// Test granting permissions
	grantMsg := types.MsgToggleTradeController{
		ChainId:          HostChainId,
		PermissionChange: types.AuthzPermissionChange_GRANT,
		Address:          tradeControllerAddress,
	}
	s.CheckICATxSubmitted(portId, channelId, func() error {
		_, err := s.GetMsgServer().ToggleTradeController(s.Ctx, &grantMsg)
		return err
	})

	// Test revoking permissions
	revokeMsg := types.MsgToggleTradeController{
		ChainId:          HostChainId,
		PermissionChange: types.AuthzPermissionChange_REVOKE,
		Address:          tradeControllerAddress,
	}
	s.CheckICATxSubmitted(portId, channelId, func() error {
		_, err := s.GetMsgServer().ToggleTradeController(s.Ctx, &revokeMsg)
		return err
	})

	// Test with an invalid chain ID - it should fail because the trade route cant be found
	invalidMsg := &types.MsgToggleTradeController{ChainId: "invalid-chain"}
	_, err := s.GetMsgServer().ToggleTradeController(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, "trade route not found")

	// Test failing to build an authz message by passing an invalid permission change
	invalidMsg = &types.MsgToggleTradeController{ChainId: HostChainId, PermissionChange: 100}
	_, err = s.GetMsgServer().ToggleTradeController(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, "invalid permission change")

	// Remove the connection ID from the trade route so the ICA submission fails
	tradeRoute.TradeAccount.ConnectionId = ""
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoute)
	_, err = s.GetMsgServer().ToggleTradeController(s.Ctx, &grantMsg)
	s.Require().ErrorContains(err, "unable to send ICA tx")
}
