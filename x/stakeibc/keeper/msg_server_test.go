package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v18/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v18/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

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
