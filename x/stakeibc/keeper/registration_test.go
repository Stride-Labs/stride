package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	epochtypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v26/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
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
	defaultRedemptionRate := sdkmath.LegacyNewDec(1)
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
		MinRedemptionRate: sdkmath.LegacyNewDec(0),
		MaxRedemptionRate: sdkmath.LegacyNewDec(0),
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
	defaultMinThreshold := sdkmath.LegacyNewDec(int64(stakeibctypes.DefaultMinRedemptionRateThreshold)).Quo(sdkmath.LegacyNewDec(100))
	defaultMaxThreshold := sdkmath.LegacyNewDec(int64(stakeibctypes.DefaultMaxRedemptionRateThreshold)).Quo(sdkmath.LegacyNewDec(100))
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
	s.Require().ErrorContains(err, "connection-id connection-10 does not exist")
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
	s.Require().ErrorContains(err, "host zone already registered for chain-id GAIA")
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
	s.Require().ErrorContains(err, "connection-id connection-1 already registered")
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
	s.Require().ErrorContains(err, "host denom uatom already registered")
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
	s.Require().ErrorContains(err, "transfer channel channel-0 already registered")
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
	s.Require().ErrorContains(err, "bech32 prefix cosmos already registered")
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
	s.Require().ErrorContains(err, "epoch unbonding record not found for epoch 3")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterDelegationAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on delegation port
	s.createActiveChannelOnICAPort("DELEGATION", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "failed to register delegation ICA")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterFeeAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on fee port
	s.createActiveChannelOnICAPort("FEE", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "failed to register fee ICA")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterWithdrawalAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on withdrawal port
	s.createActiveChannelOnICAPort("WITHDRAWAL", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "failed to register withdrawal ICA")
}

func (s *KeeperTestSuite) TestRegisterHostZone_CannotRegisterRedemptionAccount() {
	// tests for a failure if the epoch unbonding record cannot be found
	tc := s.SetupRegisterHostZone()

	// Create channel on redemption port
	s.createActiveChannelOnICAPort("REDEMPTION", "channel-1")

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().ErrorContains(err, "failed to register redemption ICA")
}

func (s *KeeperTestSuite) TestRegisterHostZone_InvalidCommunityPoolTreasuryAddress() {
	// tests for a failure if the community pool treasury address is invalid
	tc := s.SetupRegisterHostZone()

	invalidMsg := tc.validMsg
	invalidMsg.CommunityPoolTreasuryAddress = "invalid_address"

	_, err := s.GetMsgServer().RegisterHostZone(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid community pool treasury address")
}
