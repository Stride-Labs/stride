package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	"github.com/gogo/protobuf/proto" //nolint:staticcheck
)

func (s *KeeperTestSuite) TestValidateLSMLiquidStake() {
	// Create and store a valid denom trace so we can succesfully parse the LSM Token
	path := "transfer/channel-0"
	ibcDenom := s.CreateAndStoreIBCDenom(LSMTokenBaseDenom)

	// Store a second valid denom trace that will not be registered with the host zone
	invalidPath := "transfer/channel-100"
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, transfertypes.DenomTrace{
		BaseDenom: LSMTokenBaseDenom,
		Path:      invalidPath,
	})

	// Store the corresponding validator in the host zone
	hostZone := types.HostZone{
		ChainId:           HostChainId,
		TransferChannelId: ibctesting.FirstChannelID,
		Validators: []*types.Validator{
			{Address: ValAddress, SlashQueryInProgress: false},
		},
		LsmLiquidStakeEnabled: true,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Fund the user so they have sufficient balance
	liquidStaker := s.TestAccs[0]
	stakeAmount := sdk.NewInt(1_000_000)
	s.FundAccount(liquidStaker, sdk.NewCoin(ibcDenom, stakeAmount))

	// Prepare a valid message and the expected associated response
	validMsg := types.MsgLSMLiquidStake{
		Creator:          liquidStaker.String(),
		Amount:           stakeAmount,
		LsmTokenIbcDenom: ibcDenom,
	}
	expectedLSMTokenDeposit := recordstypes.LSMTokenDeposit{
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		IbcDenom:         ibcDenom,
		StakerAddress:    liquidStaker.String(),
		ValidatorAddress: ValAddress,
		Amount:           stakeAmount,
		Status:           recordstypes.LSMTokenDeposit_DEPOSIT_PENDING,
	}

	// Confirm response matches after a valid message
	lsmLiquidStake, err := s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, validMsg)
	s.Require().NoError(err, "no error expected when validating valid message")

	s.Require().Equal(HostChainId, lsmLiquidStake.HostZone.ChainId, "host zone after valid message")
	s.Require().Equal(ValAddress, lsmLiquidStake.Validator.Address, "validator after valid message")
	s.Require().Equal(expectedLSMTokenDeposit, *lsmLiquidStake.Deposit, "deposit after valid message")

	// Try with an ibc denom that's not registered - it should fail
	invalidMsg := validMsg
	invalidMsg.LsmTokenIbcDenom = transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s", path, "fake_denom")).IBCDenom()
	_, err = s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, fmt.Sprintf("denom trace not found for %s", invalidMsg.LsmTokenIbcDenom))

	// Try with a user that has insufficient balance - it should fail
	invalidMsg = validMsg
	invalidMsg.Creator = s.TestAccs[1].String()
	_, err = s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, "insufficient funds")

	// Try with with a different transfer channel - it should fail
	invalidMsg = validMsg
	invalidMsg.LsmTokenIbcDenom = transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s", invalidPath, LSMTokenBaseDenom)).IBCDenom()
	_, err = s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-100) does not match any registered host zone")

	// Flag the validator as slashed - it should fail
	hostZone.Validators[0].SlashQueryInProgress = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, err = s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, invalidMsg)
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-100) does not match any registered host zone")

	// Remove the validator and try again - it should fail
	hostZone.Validators = []*types.Validator{}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, err = s.App.StakeibcKeeper.ValidateLSMLiquidStake(s.Ctx, validMsg)
	s.Require().ErrorContains(err, fmt.Sprintf("validator (%s) is not registered in the Stride validator set", ValAddress))
}

func (s *KeeperTestSuite) TestGetLSMTokenDenomTrace() {
	baseDenom := "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/48"
	path := "transfer/channel-0"
	ibcDenom := transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s", path, baseDenom)).IBCDenom()

	// Store denom trace so the transfer keeper can look it up
	expectedDenomTrace := transfertypes.DenomTrace{
		BaseDenom: baseDenom,
		Path:      path,
	}
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, expectedDenomTrace)

	// Test parsing of IBC Denom
	actualDenomTrace, err := s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, ibcDenom)
	s.Require().NoError(err, "no error expected with successful parse")
	s.Require().Equal(expectedDenomTrace, actualDenomTrace, "denom trace")

	// Attempt to parse with a non-ibc denom - it should fail
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, "non-ibc-denom")
	s.Require().ErrorContains(err, "lsm token is not an IBC token (non-ibc-denom)")

	// Attempt to parse with an invalid ibc-denom - it should fail
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, "ibc/xxx")
	s.Require().ErrorContains(err, "unable to get ibc hex hash from denom ibc/xxx")

	// Attempt to parse with a valid ibc denom that is not registered - it should fail
	notRegisteredIBCDenom := transfertypes.ParseDenomTrace("transfer/channel-0/cosmosXXX").IBCDenom()
	_, err = s.App.StakeibcKeeper.GetLSMTokenDenomTrace(s.Ctx, notRegisteredIBCDenom)
	s.Require().ErrorContains(err, "denom trace not found")
}

func (s *KeeperTestSuite) TestIsValidIBCPath() {
	validIBCPaths := []string{
		"transfer/channel-0",
		"transfer/channel-10",
		"transfer/channel-99999",
	}
	invalidIBCPaths := []string{
		"transferx/channel-0",
		"transfer/channel-X",
		"transfer/channel-0/transfer/channel-1",
	}

	for _, validPath := range validIBCPaths {
		s.Require().True(keeper.IsValidIBCPath(validPath), "should be valid")
	}
	for _, validPath := range invalidIBCPaths {
		s.Require().False(keeper.IsValidIBCPath(validPath), "should be invalid")
	}
}

func (s *KeeperTestSuite) TestGetHostZoneFromLSMTokenPath() {
	// Set a host zone in the store with channel-0
	hostZone := types.HostZone{
		ChainId:               HostChainId,
		TransferChannelId:     ibctesting.FirstChannelID,
		LsmLiquidStakeEnabled: true,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Successful lookup
	validPath := fmt.Sprintf("%s/%s", transfertypes.PortID, ibctesting.FirstChannelID)
	hostZone, err := s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, validPath)
	s.Require().NoError(err, "no error expected from valid path")
	s.Require().Equal(HostChainId, hostZone.ChainId, "host zone")

	// Invalid IBC path should fail
	_, err = s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, "transfer/channel-0/transfer/channel-1")
	s.Require().ErrorContains(err, "ibc path of LSM token (transfer/channel-0/transfer/channel-1) cannot be more than 1 hop away")

	// Passing an unregistered channel-id should cause it to fail
	_, err = s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, "transfer/channel-1")
	s.Require().ErrorContains(err, "transfer channel-id from LSM token (channel-1) does not match any registered host zone")

	// Disabling LSM for the host should cause it to fail
	hostZone.LsmLiquidStakeEnabled = false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	_, err = s.App.StakeibcKeeper.GetHostZoneFromLSMTokenPath(s.Ctx, validPath)
	s.Require().ErrorContains(err, "LSM liquid stake disabled for GAIA")
}

func (s *KeeperTestSuite) TestGetValidatorFromLSMTokenDenom() {
	valAddress := "cosmosvaloperXXX"
	denom := valAddress + "/42" // add record ID
	validators := []*types.Validator{{Address: valAddress, SlashQueryInProgress: false}}

	// Successful lookup
	validator, err := s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom(denom, validators)
	s.Require().NoError(err, "no error expected from valid lsm denom")
	s.Require().Equal(valAddress, validator.Address, "host zone")

	// Invalid LSM denoms - should fail
	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom("invalid_denom", validators)
	s.Require().ErrorContains(err, "lsm token base denom is not of the format {val-address}/{record-id} (invalid_denom)")

	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom("cosmosvaloperXXX/42/1", validators)
	s.Require().ErrorContains(err, "lsm token base denom is not of the format {val-address}/{record-id} (cosmosvaloperXXX/42/1)")

	// Validator does not exist - should fail
	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom(denom, []*types.Validator{})
	s.Require().ErrorContains(err, "validator (cosmosvaloperXXX) is not registered in the Stride validator set")

	// Pass in a validator that has a slash query in flight - it should fail
	validatorsWithPendingQuery := []*types.Validator{{Address: valAddress, SlashQueryInProgress: true}}
	_, err = s.App.StakeibcKeeper.GetValidatorFromLSMTokenDenom(denom, validatorsWithPendingQuery)
	s.Require().ErrorContains(err, "validator cosmosvaloperXXX was slashed")
}

func (s *KeeperTestSuite) TestShouldCheckIfValidatorWasSlashed() {
	testCases := []struct {
		name                string
		checkpoint          sdkmath.Int
		progress            sdkmath.Int
		stakeAmount         sdkmath.Int
		expectedShouldQuery bool
	}{
		{
			// Checkpoint: 1000, Stake: 99
			// Old Progress: 900, New Progress: 900 + 99 = 999
			// Old Interval: 900 / 1000 = Interval #0
			// New Interval: 999 / 1000 = Interval #0 (no query)
			name:                "case #1 - short of checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(99),
			expectedShouldQuery: false,
		},
		{
			// Checkpoint: 1000, Stake: 100
			// Old Progress: 900, New Progress: 900 + 100 = 1000
			// Old Interval: 900 / 1000 = Interval #0
			// New Interval: 1000 / 1000 = Interval #1 (query)
			name:                "case #1 - at checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(100),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint: 1000, Stake: 101
			// Old Progress: 900, New Progress: 900 + 101 = 1000
			// Old Interval: 900 / 1000 = Interval #0
			// New Interval: 1001 / 1000 = Interval #1 (query)
			name:                "case #1 - past checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(101),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint: 1000, Stake: 99
			// Old Progress: 11,900, New Progress: 11,900 + 99 = 11,999
			// Old Interval: 11,900 / 1000 = Interval #11
			// New Interval: 11,999 / 1000 = Interval #11 (query)
			name:                "case #2 - short of checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(11_900),
			stakeAmount:         sdk.NewInt(99),
			expectedShouldQuery: false,
		},
		{
			// Checkpoint: 1000, Stake: 100
			// Old Progress: 11,900, New Progress: 11,900 + 100 = 12,000
			// Old Interval: 11,900 / 1000 = Interval #11
			// New Interval: 12,000 / 1000 = Interval #12 (query)
			name:                "case #2 - at checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(11_900),
			stakeAmount:         sdk.NewInt(100),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint: 1000, Stake: 101
			// Old Progress: 11,900, New Progress: 11,900 + 101 = 12,001
			// Old Interval: 11,900 / 1000 = Interval #11
			// New Interval: 12,001 / 1000 = Interval #12 (query)
			name:                "case #2 - past checkpoint",
			checkpoint:          sdkmath.NewInt(1000),
			progress:            sdk.NewInt(11_900),
			stakeAmount:         sdk.NewInt(101),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint: 6,890, Stake: 339
			// Old Progress: 41,000, New Progress: 41,000 + 339 = 41,339
			// Old Interval: 41,000 / 6,890 = Interval #5
			// New Interval: 41,339 / 6,890 = Interval #5 (no query)
			name:                "case #3 - short of checkpoint",
			checkpoint:          sdkmath.NewInt(6890),
			progress:            sdk.NewInt(41_000),
			stakeAmount:         sdk.NewInt(101),
			expectedShouldQuery: false,
		},
		{
			// Checkpoint: 6,890, Stake: 340
			// Old Progress: 41,000, New Progress: 41,000 + 440 = 41,440
			// Old Interval: 41,000 / 6,890 = Interval #5
			// New Interval: 41,440 / 6,890 = Interval #6 (query)
			name:                "case #3 - at checkpoint",
			checkpoint:          sdkmath.NewInt(6890),
			progress:            sdk.NewInt(41_000),
			stakeAmount:         sdk.NewInt(340),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint: 6,890
			// Old Progress: 41,000, New Progress: 41,000 + 441 = 41,440
			// Old Interval: 41,000 / 6,890 = Interval #5
			// New Interval: 41,441 / 6,890 = Interval #6 (query)
			name:                "case #3 - past checkpoint",
			checkpoint:          sdkmath.NewInt(6890),
			progress:            sdk.NewInt(41_000),
			stakeAmount:         sdk.NewInt(341),
			expectedShouldQuery: true,
		},
		{
			// Checkpoint of 0 - should not issue query
			name:                "threshold of 0",
			checkpoint:          sdkmath.ZeroInt(),
			progress:            sdk.NewInt(41_000),
			stakeAmount:         sdk.NewInt(340),
			expectedShouldQuery: false,
		},
	}

	for _, tc := range testCases {
		// Store query interval param
		validator := types.Validator{SlashQueryProgressTracker: tc.progress, SlashQueryCheckpoint: tc.checkpoint}
		actualShouldQuery := s.App.StakeibcKeeper.ShouldCheckIfValidatorWasSlashed(s.Ctx, validator, tc.stakeAmount)
		s.Require().Equal(tc.expectedShouldQuery, actualShouldQuery, tc.name)
	}
}

func (s *KeeperTestSuite) TestGetUpdatedSlashQueryCheckpoint() {
	testCases := []struct {
		name               string
		threshold          uint64
		totalDelegations   sdkmath.Int
		expectedCheckpoint sdkmath.Int
	}{
		{
			name:               "10%",
			threshold:          10,
			totalDelegations:   sdkmath.NewInt(1_000_000),
			expectedCheckpoint: sdkmath.NewInt(100_000),
		},
		{
			name:               "25%",
			threshold:          25,
			totalDelegations:   sdkmath.NewInt(1_000_000),
			expectedCheckpoint: sdkmath.NewInt(250_000),
		},
		{
			name:               "75%",
			threshold:          75,
			totalDelegations:   sdkmath.NewInt(1_000_000),
			expectedCheckpoint: sdkmath.NewInt(750_000),
		},
		{
			name:               "int truncation",
			threshold:          10,
			totalDelegations:   sdkmath.NewInt(39),
			expectedCheckpoint: sdkmath.NewInt(3),
		},
		{
			name:               "0-TVL",
			threshold:          10,
			totalDelegations:   sdkmath.ZeroInt(),
			expectedCheckpoint: sdkmath.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Set the slash query threshold
			params := s.App.StakeibcKeeper.GetParams(s.Ctx)
			params.ValidatorSlashQueryThreshold = tc.threshold
			s.App.StakeibcKeeper.SetParams(s.Ctx, params)

			// Check the new checkpoint
			actualCheckpoint := s.App.StakeibcKeeper.GetUpdatedSlashQueryCheckpoint(s.Ctx, tc.totalDelegations)
			s.Require().Equal(tc.expectedCheckpoint.Int64(), actualCheckpoint.Int64(), "checkpoint")
		})
	}
}

func (s *KeeperTestSuite) TestTransferAllLSMDeposits() {
	s.CreateTransferChannel(HostChainId)

	// Create a valid IBC denom
	ibcDenom := s.CreateAndStoreIBCDenom(LSMTokenBaseDenom)

	// Store 2 host zones - one that was registered successfully,
	// and one that's missing a delegation channel
	hostZones := []types.HostZone{
		{
			// Valid host zone
			ChainId:              HostChainId,
			TransferChannelId:    ibctesting.FirstChannelID,
			DepositAddress:       s.TestAccs[1].String(),
			DelegationIcaAddress: DelegationICAAddress,
		},
		{
			// Missing delegation ICA
			ChainId:           "chain-2",
			TransferChannelId: "channel-2",
			DepositAddress:    "stride_DEPOSIT_2",
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// For each host chain store 4 deposits:
	//   - One ready to be transferred with a valid IBC denom
	//   - One ready to be transferred with an invalid IBC denom (should fail)
	//   - One not ready to be transferred with a valid IBC denom
	//   - One not ready to be transferred with an invalid IBC denom
	expectedDepositStatus := map[string]recordstypes.LSMTokenDeposit_Status{}
	for _, chainId := range []string{HostChainId, OsmoChainId} {
		for _, startingStatus := range []recordstypes.LSMTokenDeposit_Status{
			recordstypes.LSMTokenDeposit_TRANSFER_QUEUE,
			recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		} {

			for i, shouldSucceed := range []bool{true, false} {
				denom := fmt.Sprintf("denom-starting-in-status-%s-%d", startingStatus.String(), i)
				depositKey := fmt.Sprintf("%s-%s", chainId, denom)

				if !shouldSucceed {
					ibcDenom = "ibc/fake_denom"
				}
				deposit := recordstypes.LSMTokenDeposit{
					ChainId:  chainId,
					Denom:    denom,
					IbcDenom: ibcDenom,
					Status:   startingStatus,
				}
				s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

				// The status should update to IN_PROGRESS if the record was queued for transfer, on the
				//   valid host zone, with a valid IBC denom
				// The status should update to FAILED if the record was queued for transfer, on the
				//   valid host zone, with an invalid IBC denom
				// The status should not change on the invalid host zone
				expectedStatus := startingStatus
				if chainId == HostChainId && startingStatus == recordstypes.LSMTokenDeposit_TRANSFER_QUEUE {
					if shouldSucceed {
						expectedStatus = recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS
					} else {
						expectedStatus = recordstypes.LSMTokenDeposit_TRANSFER_FAILED
					}
				}

				expectedDepositStatus[depositKey] = expectedStatus
			}
		}
	}

	// Call transfer across all hosts
	s.App.StakeibcKeeper.TransferAllLSMDeposits(s.Ctx)

	// Check that the status of the relevant records was updated
	allDeposits := s.App.RecordsKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Len(allDeposits, 8) // 4 host zones, 2 statuses, 2 deposits = 2 * 2 * 2 = 8

	for _, deposit := range allDeposits {
		depositKey := fmt.Sprintf("%s-%s", deposit.ChainId, deposit.Denom)
		s.Require().Equal(expectedDepositStatus[depositKey].String(), deposit.Status.String(), "deposit status for %s", depositKey)
	}
}

func (s *KeeperTestSuite) TestDetokenizeLSMDeposit() {
	// Create the delegation ICA
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	s.CreateICAChannel(owner)
	portId, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "no error expected when formatting portId")

	// Get the ica address that was just created
	delegationICAAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portId)
	s.Require().True(found, "ICA account should have been created")
	s.Require().NotEmpty(delegationICAAddress, "ICA Address should not be empty")

	// Build the host zone and deposit (which are arguments to detokenize)
	initialHostZone := types.HostZone{
		ChainId:              HostChainId,
		DelegationIcaAddress: delegationICAAddress,
		ConnectionId:         ibctesting.FirstConnectionID,
		Validators:           []*types.Validator{{DelegationChangesInProgress: 0}},
	}

	denom := "cosmosvalXXX/42"
	initalDeposit := recordstypes.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   denom,
		Amount:  sdk.NewInt(1000),
		Status:  recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
		StToken: sdk.NewCoin(StAtom, sdk.OneInt()),
	}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, initalDeposit)

	// Successfully Detokenize
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, initialHostZone, initalDeposit)
	s.Require().NoError(err, "no error expected when detokenizing")

	// Confirm deposit status was updated
	finalDeposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, denom)
	s.Require().True(found, "deposit should have been found")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS.String(), finalDeposit.Status.String(), "deposit status")

	// Check callback data was stored
	allCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(allCallbackData, 1, "length of callback data")

	var callbackData types.DetokenizeSharesCallback
	err = proto.Unmarshal(allCallbackData[0].CallbackArgs, &callbackData)
	s.Require().NoError(err, "no error expected when unmarshalling callback data")

	s.Require().Equal(initalDeposit, *callbackData.Deposit, "callback data LSM deposit")

	// Check the number of delegation changes was incremented
	finalHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(1, int(finalHostZone.Validators[0].DelegationChangesInProgress), "delegation changes in progress")

	// Remove connection ID and re-submit - should fail
	hostZoneWithoutConnectionId := initialHostZone
	hostZoneWithoutConnectionId.ConnectionId = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZoneWithoutConnectionId, initalDeposit)
	s.Require().ErrorContains(err, "unable to submit detokenization ICA")

	// Remove delegation account and re-submit - should also fail
	hostZoneWithoutDelegationAccount := initialHostZone
	hostZoneWithoutDelegationAccount.DelegationIcaAddress = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZoneWithoutDelegationAccount, initalDeposit)
	s.Require().ErrorContains(err, "no delegation account found")
}

func (s *KeeperTestSuite) TestDetokenizeAllLSMDeposits() {
	// Create an open delegation ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_DELEGATION)
	s.CreateICAChannel(owner)
	portId, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "no error expected when formatting portId")

	// Get the ica address that was just created
	delegationICAAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portId)
	s.Require().True(found, "ICA account should have been created")
	s.Require().NotEmpty(delegationICAAddress, "ICA Address should not be empty")

	// Store two host zones - one with an open Delegation channel, and one without
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		DelegationIcaAddress: delegationICAAddress,
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:      OsmoChainId,
		ConnectionId: "connection-2",
	})

	// For each host chain store 4 deposits
	// 2 of which are ready to be detokenized, and 2 of which are not
	expectedDepositStatus := map[string]recordstypes.LSMTokenDeposit_Status{}
	for _, chainId := range []string{HostChainId, OsmoChainId} {
		for _, startingStatus := range []recordstypes.LSMTokenDeposit_Status{
			recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE,
			recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS,
		} {

			for i := 0; i < 2; i++ {
				denom := fmt.Sprintf("denom-starting-in-status-%s-%d", startingStatus.String(), i)
				depositKey := fmt.Sprintf("%s-%s", chainId, denom)
				deposit := recordstypes.LSMTokenDeposit{
					ChainId: chainId,
					Denom:   denom,
					Status:  startingStatus,
				}
				s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

				// The status is only expected to change for the QUEUED records on the
				// host chain with the open delegation channel
				expectedStatus := startingStatus
				if chainId == HostChainId && startingStatus == recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE {
					expectedStatus = recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS
				}
				expectedDepositStatus[depositKey] = expectedStatus
			}
		}
	}

	// Call detokenization across all hosts
	s.App.StakeibcKeeper.DetokenizeAllLSMDeposits(s.Ctx)

	// Check that the status of the relevant records was updated
	allDeposits := s.App.RecordsKeeper.GetAllLSMTokenDeposit(s.Ctx)
	s.Require().Len(allDeposits, 8) // 2 host zones, 2 statuses, 2 deposits = 2 * 2 * 2 = 8

	for _, deposit := range allDeposits {
		depositKey := fmt.Sprintf("%s-%s", deposit.ChainId, deposit.Denom)
		s.Require().Equal(expectedDepositStatus[depositKey].String(), deposit.Status.String(), "deposit status for %s", depositKey)
	}
}
