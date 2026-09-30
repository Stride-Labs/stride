package keeper_test

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	recordtypes "github.com/Stride-Labs/stride/v34/x/records/types"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

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
	_, err := s.GetMsgServer().UpdateHostZoneParams(s.Ctx, &validUpdateMsg)
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
	_, err = s.GetMsgServer().UpdateHostZoneParams(s.Ctx, &validUpdateMsg)
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
	_, err = s.GetMsgServer().UpdateHostZoneParams(s.Ctx, &invalidUpdateMsg)
	s.Require().ErrorContains(err, "host zone not found")

	// Finally attempt again with an invalid authority, it should also fail
	invalidUpdateMsg = types.MsgUpdateHostZoneParams{
		Authority:           "invalid-authority",
		ChainId:             HostChainId,
		MaxMessagesPerIcaTx: updatedMessages,
	}
	_, err = s.GetMsgServer().UpdateHostZoneParams(s.Ctx, &invalidUpdateMsg)
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
		validator.SharesToTokensRate = sdkmath.LegacyOneDec()
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &tc.validMsg)
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &badHostZoneMsg)
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &tc.validMsg)
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &tc.validMsg)
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &addValidatorMsg)
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
	_, err := s.GetMsgServer().AddValidators(s.Ctx, &addValidatorMsg)
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
			SharesToTokensRate: sdkmath.LegacyOneDec(),
		},
		{
			Name:               "val2",
			Address:            "stride_VAL2",
			Weight:             0,
			Delegation:         sdkmath.ZeroInt(),
			SharesToTokensRate: sdkmath.LegacyOneDec(),
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
	_, err := s.GetMsgServer().DeleteValidator(s.Ctx, &tc.validMsgs[0])
	s.Require().NoError(err)

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")
	s.Require().Equal(1, len(hostZone.Validators), "number of validators should be 1")
	s.Require().Equal(tc.initialValidators[1:], hostZone.Validators, "validators list after removing 1 validator")

	// Delete second validator
	_, err = s.GetMsgServer().DeleteValidator(s.Ctx, &tc.validMsgs[1])
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
	_, err := s.GetMsgServer().DeleteValidator(s.Ctx, &badHostZoneMsg)
	s.Require().ErrorContains(err, "host zone gaia not found")
}

func (s *KeeperTestSuite) TestDeleteValidator_AddressNotFound() {
	tc := s.SetupDeleteValidator()

	// Build message with a validator address that does not exist
	badAddressMsg := tc.validMsgs[0]
	badAddressMsg.ValAddr = "stride_VAL5"
	_, err := s.GetMsgServer().DeleteValidator(s.Ctx, &badAddressMsg)

	s.Require().ErrorContains(err, "failed to remove validator stride_VAL5 from host zone GAIA")
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroDelegation() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero delegation
	hostZone := tc.hostZone
	hostZone.Validators[0].Delegation = sdkmath.NewInt(1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(s.Ctx, &tc.validMsgs[0])
	s.Require().ErrorContains(err, "Validator (stride_VAL1) has non-zero delegation (1) or weight (0)")
}

func (s *KeeperTestSuite) TestDeleteValidator_NonZeroWeight() {
	tc := s.SetupDeleteValidator()

	// Update val1 to have a non-zero weight
	hostZone := tc.hostZone
	hostZone.Validators[0].Weight = 1
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.GetMsgServer().DeleteValidator(s.Ctx, &tc.validMsgs[0])
	s.Require().ErrorContains(err, "Validator (stride_VAL1) has non-zero delegation (0) or weight (1)")
}

// ----------------------------------------------------
//	                 Account
// ----------------------------------------------------

type Account struct {
	acc           sdk.AccAddress
	atomBalance   sdk.Coin
	stAtomBalance sdk.Coin
}

// ----------------------------------------------------
//	           RestoreInterchainAccount
// ----------------------------------------------------

type DepositRecordStatusUpdate struct {
	chainId                         string
	initialStatus                   recordtypes.DepositRecord_Status
	revertedStatus                  recordtypes.DepositRecord_Status
	initialDelegationTxsInProgress  uint64
	revertedDelegationTxsInProgress uint64
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
		RedemptionRate: sdkmath.LegacyOneDec(), // if not set, the beginblocker invariant panics
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
			chainId:                         HostChainId,
			initialStatus:                   recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
			revertedStatus:                  recordtypes.DepositRecord_TRANSFER_IN_PROGRESS,
			initialDelegationTxsInProgress:  2,
			revertedDelegationTxsInProgress: 2,
		},
		{
			// Status gets reverted from IN_PROGRESS to QUEUE
			chainId:                         HostChainId,
			initialStatus:                   recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus:                  recordtypes.DepositRecord_DELEGATION_QUEUE,
			initialDelegationTxsInProgress:  2,
			revertedDelegationTxsInProgress: 0,
		},
		{
			// Status doesn't get reveted because it's a different host zone
			chainId:                         "different_host_zone",
			initialStatus:                   recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			revertedStatus:                  recordtypes.DepositRecord_DELEGATION_IN_PROGRESS,
			initialDelegationTxsInProgress:  2,
			revertedDelegationTxsInProgress: 2,
		},
	}
	for i, depositRecord := range depositRecords {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordtypes.DepositRecord{
			Id:                      uint64(i),
			HostZoneId:              depositRecord.chainId,
			Status:                  depositRecord.initialStatus,
			DelegationTxsInProgress: depositRecord.initialDelegationTxsInProgress,
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
func (s *KeeperTestSuite) restoreChannelAndVerifySuccess(msg types.MsgRestoreInterchainAccount, portID, channelID string) {
	// Restore the channel
	_, err := s.GetMsgServer().RestoreInterchainAccount(s.Ctx, &msg)
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
// and the delegation txs in progress was set to 0 on each deposit record
func (s *KeeperTestSuite) verifyDelegationChangeInProgressReset(expectedDepositRecords []DepositRecordStatusUpdate) {
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Len(hostZone.Validators, 3, "there should be 3 validators on this host zone")

	for _, validator := range hostZone.Validators {
		s.Require().Zero(validator.DelegationChangesInProgress,
			"delegation change in progress should have been reset for validator %s", validator.Address)
	}

	for i, expectedRecord := range expectedDepositRecords {
		actualRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, uint64(i))
		s.Require().True(found, "deposit record %d should have been found", i)
		s.Require().Equal(expectedRecord.revertedDelegationTxsInProgress, actualRecord.DelegationTxsInProgress,
			"delegation txs in progress fro record %d", i)
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

	// Close the delegation channel with a pending undelegation batch stranded on it
	s.closeICAChannel(tc.delegationPortID, tc.delegationChannelID)
	s.App.StakeibcKeeper.SetPendingUndelegation(s.Ctx, HostChainId, sdkmath.NewInt(500))
	s.App.StakeibcKeeper.SetPendingUndelegationInFlight(s.Ctx, HostChainId, 2)

	// Confirm the new channel was created
	s.restoreChannelAndVerifySuccess(tc.validMsg, tc.delegationPortID, tc.delegationChannelID)

	// Verify the record status' were reverted
	s.verifyDepositRecordsStatus(tc.depositRecordStatusUpdates, true)
	s.verifyHostZoneUnbondingStatus(tc.unbondingRecordStatusUpdate, true)
	s.verifyLSMDepositStatus(tc.lsmTokenDepositStatusUpdate, true)
	s.verifyDelegationChangeInProgressReset(tc.depositRecordStatusUpdates)
	s.verifyUndelegationChangeInProgressReset()

	// The stranded batch is released for the next day epoch; the amount itself is untouched
	s.Require().Zero(s.App.StakeibcKeeper.GetPendingUndelegationInFlight(s.Ctx, HostChainId), "in-flight batches cleared by restore")
	pending, found := s.App.StakeibcKeeper.GetPendingUndelegation(s.Ctx, HostChainId)
	s.Require().True(found, "pending undelegation kept by restore")
	s.Require().Equal(sdkmath.NewInt(500), pending, "pending amount unchanged by restore")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_InvalidConnectionId() {
	tc := s.SetupRestoreInterchainAccount(false)

	// Update the connectionId on the host zone so that it doesn't exist
	invalidMsg := tc.validMsg
	invalidMsg.ConnectionId = "fake_connection"

	_, err := s.GetMsgServer().RestoreInterchainAccount(s.Ctx, &invalidMsg)
	s.Require().ErrorContains(err, "connection fake_connection not found")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_CannotRestoreNonExistentAcct() {
	tc := s.SetupRestoreInterchainAccount(false)

	// Attempt to restore an account that does not exist
	msg := tc.validMsg
	msg.AccountOwner = types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_WITHDRAWAL)

	_, err := s.GetMsgServer().RestoreInterchainAccount(s.Ctx, &msg)
	s.Require().ErrorContains(err, "ICA controller account address not found: GAIA.WITHDRAWAL")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_HostZoneNotFound() {
	tc := s.SetupRestoreInterchainAccount(true)
	s.closeICAChannel(tc.delegationPortID, tc.delegationChannelID)

	// Delete the host zone so the lookup fails
	// (this check only runs for the delegation channel)
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	_, err := s.GetMsgServer().RestoreInterchainAccount(s.Ctx, &tc.validMsg)
	s.Require().ErrorContains(err, "delegation ICA supplied, but no associated host zone")
}

func (s *KeeperTestSuite) TestRestoreInterchainAccount_RevertDepositRecords_Failure() {
	tc := s.SetupRestoreInterchainAccount(true)

	_, err := s.GetMsgServer().RestoreInterchainAccount(s.Ctx, &tc.validMsg)
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
		RedemptionRate:    sdkmath.LegacyNewDec(1.0),
		MinRedemptionRate: sdkmath.LegacyNewDec(9).Quo(sdkmath.LegacyNewDec(10)),
		MaxRedemptionRate: sdkmath.LegacyNewDec(15).Quo(sdkmath.LegacyNewDec(10)),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	defaultMsg := stakeibctypes.MsgUpdateInnerRedemptionRateBounds{
		// TODO: does this need to be the admin address?
		Creator:                s.TestAccs[0].String(),
		ChainId:                HostChainId,
		MinInnerRedemptionRate: sdkmath.LegacyNewDec(1),
		MaxInnerRedemptionRate: sdkmath.LegacyNewDec(11).Quo(sdkmath.LegacyNewDec(10)),
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
	tc.validMsg.MinInnerRedemptionRate = sdkmath.LegacyNewDec(0)

	// Set the inner bounds on the host zone
	_, err := s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg := fmt.Sprintf("inner min safety threshold (%s) is less than outer min safety threshold (%s)", tc.validMsg.MinInnerRedemptionRate, sdkmath.LegacyNewDec(9).Quo(sdkmath.LegacyNewDec(10)))
	s.Require().ErrorContains(err, errMsg)

	// Set the min inner bound to be valid, but the max inner bound to be greater than the max outer bound
	tc.validMsg.MinInnerRedemptionRate = sdkmath.LegacyNewDec(1)
	tc.validMsg.MaxInnerRedemptionRate = sdkmath.LegacyNewDec(3)
	// Set the inner bounds on the host zone
	_, err = s.GetMsgServer().UpdateInnerRedemptionRateBounds(s.Ctx, &tc.validMsg)
	// verify it throws an error
	errMsg = fmt.Sprintf("inner max safety threshold (%s) is greater than outer max safety threshold (%s)", tc.validMsg.MaxInnerRedemptionRate, sdkmath.LegacyNewDec(15).Quo(sdkmath.LegacyNewDec(10)))
	s.Require().ErrorContains(err, errMsg)
}

// Validate basic tests
func (s *KeeperTestSuite) TestUpdateInnerRedemptionRateBounds_InvalidMsg() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the min inner bound to be greater than than the max inner bound
	invalidMsg := tc.validMsg
	invalidMsg.MinInnerRedemptionRate = sdkmath.LegacyNewDec(2)

	err := invalidMsg.ValidateBasic()

	// Verify the error
	errMsg := fmt.Sprintf("Inner max safety threshold (%s) is less than inner min safety threshold (%s)", invalidMsg.MaxInnerRedemptionRate, invalidMsg.MinInnerRedemptionRate)
	s.Require().ErrorContains(err, errMsg)
}

// Verify that if inner bounds end up outside of outer bounds (somehow), the outer bounds are returned
func (s *KeeperTestSuite) TestGetInnerSafetyBounds() {
	tc := s.SetupUpdateInnerRedemptionRateBounds()

	// Set the inner bounds outside the outer bounds on the host zone directly
	tc.zone.MinInnerRedemptionRate = sdkmath.LegacyNewDec(0)
	tc.zone.MaxInnerRedemptionRate = sdkmath.LegacyNewDec(3)
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
		RedemptionRate:    sdkmath.LegacyNewDec(1.0),
		MinRedemptionRate: sdkmath.LegacyNewDec(9).Quo(sdkmath.LegacyNewDec(10)),
		MaxRedemptionRate: sdkmath.LegacyNewDec(15).Quo(sdkmath.LegacyNewDec(10)),
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
	s.Require().ErrorContains(err, "host zone invalid-chain not found")
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
	s.Require().Error(err, "host zone GAIA is not halted")
}

// ----------------------------------------------------
//	              CalibrateDelegation
// ----------------------------------------------------

// Sets up a host zone with a delegation ICA (so the calibration ICQ can be submitted) and two validators,
// where the queried validator has a stale in-flight counter
func (s *KeeperTestSuite) SetupCalibrateDelegation() types.MsgCalibrateDelegation {
	s.CreateTransferChannel(HostChainId)

	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		DelegationIcaAddress: s.IcaAddresses[delegationAccountOwner],
		TotalDelegations:     sdkmath.NewInt(1_000_000),
		Validators: []*types.Validator{
			{Address: "valoper1", DelegationChangesInProgress: 3},
			{
				Address:                     ValAddress,
				Delegation:                  sdkmath.NewInt(10_000),
				SharesToTokensRate:          sdkmath.LegacyMustNewDecFromStr("0.75"),
				DelegationChangesInProgress: 2,
			},
		},
	})

	return types.MsgCalibrateDelegation{
		Creator: "creator",
		ChainId: HostChainId,
		Valoper: ValAddress,
	}
}

func (s *KeeperTestSuite) TestCalibrateDelegation_ResetFlag() {
	msg := s.SetupCalibrateDelegation()
	msg.ResetDelegationChangesInProgress = true

	_, err := s.GetMsgServer().CalibrateDelegation(s.Ctx, &msg)
	s.Require().NoError(err)

	// Only the queried validator's flag is zeroed
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(3), hostZone.Validators[0].DelegationChangesInProgress, "other validator untouched")
	s.Require().Equal(int64(0), hostZone.Validators[1].DelegationChangesInProgress, "queried validator reset")
	s.Require().Equal(int64(10_000), hostZone.Validators[1].Delegation.Int64(), "delegation unchanged by the reset")

	s.Require().Len(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), 1, "calibration query submitted")
}

func (s *KeeperTestSuite) TestCalibrateDelegation_QueryOptsIntoEmptyResponse() {
	msg := s.SetupCalibrateDelegation()

	_, err := s.GetMsgServer().CalibrateDelegation(s.Ctx, &msg)
	s.Require().NoError(err)

	queries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(queries, 1, "calibration query submitted")
	s.Require().True(queries[0].InvokeCallbackOnEmptyResponse, "query opts into empty responses")
	s.Require().Equal(ValAddress, string(queries[0].CallbackData), "callback data is the validator address")
}

func (s *KeeperTestSuite) TestCalibrateDelegation_NoReset() {
	msg := s.SetupCalibrateDelegation()

	_, err := s.GetMsgServer().CalibrateDelegation(s.Ctx, &msg)
	s.Require().NoError(err)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(3), hostZone.Validators[0].DelegationChangesInProgress, "other validator untouched")
	s.Require().Equal(int64(2), hostZone.Validators[1].DelegationChangesInProgress, "queried validator untouched")

	s.Require().Len(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), 1, "calibration query submitted")
}

func (s *KeeperTestSuite) TestCalibrateDelegation_ResetUnknownValidator() {
	msg := s.SetupCalibrateDelegation()
	msg.ResetDelegationChangesInProgress = true
	msg.Valoper = "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne"

	_, err := s.GetMsgServer().CalibrateDelegation(s.Ctx, &msg)
	s.Require().ErrorIs(err, types.ErrValidatorNotFound)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(3), hostZone.Validators[0].DelegationChangesInProgress)
	s.Require().Equal(int64(2), hostZone.Validators[1].DelegationChangesInProgress)
	s.Require().Empty(s.App.InterchainqueryKeeper.AllQueries(s.Ctx), "no query submitted")
}

// A stale flag blocks the callback forever; a calibrate with the reset lets the next callback apply the correction
func (s *KeeperTestSuite) TestCalibrateDelegation_StaleFlagBlocksUntilReset() {
	msg := s.SetupCalibrateDelegation()
	query := icqtypes.Query{ChainId: HostChainId}

	// 20,000 shares * 0.75 = 15,000 tokens, versus the 10,000 recorded
	queryResponse := s.CreateDelegatorSharesQueryResponse(ValAddress, sdkmath.LegacyMustNewDecFromStr("20000"))

	// Without the reset, the callback is a no-op
	err := keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
	s.Require().NoError(err)
	s.Require().Equal(int64(10_000), s.MustGetHostZone(HostChainId).Validators[1].Delegation.Int64(), "blocked by stale flag")

	// Calibrate with the reset, then the callback applies the correction
	msg.ResetDelegationChangesInProgress = true
	_, err = s.GetMsgServer().CalibrateDelegation(s.Ctx, &msg)
	s.Require().NoError(err)

	err = keeper.CalibrateDelegationCallback(s.App.StakeibcKeeper, s.Ctx, queryResponse, query)
	s.Require().NoError(err)

	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(int64(15_000), hostZone.Validators[1].Delegation.Int64(), "correction applied")
	s.Require().Equal(int64(1_005_000), hostZone.TotalDelegations.Int64(), "total delegation corrected")
}
