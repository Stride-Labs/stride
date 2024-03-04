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

// ----------------------------------------------------
//	                   LiquidStake
// ----------------------------------------------------

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
