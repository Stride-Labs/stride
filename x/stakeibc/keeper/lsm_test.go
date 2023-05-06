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
	ibcDenom := transfertypes.ParseDenomTrace(fmt.Sprintf("%s/%s", path, LSMTokenBaseDenom)).IBCDenom()
	expectedDenomTrace := transfertypes.DenomTrace{
		BaseDenom: LSMTokenBaseDenom,
		Path:      path,
	}
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, expectedDenomTrace)

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
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:           HostChainId,
		TransferChannelId: ibctesting.FirstChannelID,
	})

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
		queryInterval       uint64
		progress            sdkmath.Int
		stakeAmount         sdkmath.Int
		expectedShouldQuery bool
	}{
		{
			name:                "interval 1 - short of checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(99),
			expectedShouldQuery: false,
		},
		{
			name:                "interval 1 - at checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(100),
			expectedShouldQuery: true,
		},
		{
			name:                "interval 1 - past checkpoint",
			queryInterval:       1000,
			progress:            sdk.NewInt(900),
			stakeAmount:         sdk.NewInt(101),
			expectedShouldQuery: true,
		},
		{
			name:                "interval 2 - short of checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(133),  // 4,133
			expectedShouldQuery: false,
		},
		{
			name:                "interval 2 - at checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(134),  // 4,134
			expectedShouldQuery: true,
		},
		{
			name:                "interval 2 - past checkpoint",
			queryInterval:       689,
			progress:            sdk.NewInt(4000), // 4,134 is checkpoint (689 * 5)
			stakeAmount:         sdk.NewInt(135),  // 4,135
			expectedShouldQuery: true,
		},
	}

	for _, tc := range testCases {
		// Store query interval param
		params := types.DefaultParams()
		params.ValidatorSlashQueryInterval = tc.queryInterval
		s.App.StakeibcKeeper.SetParams(s.Ctx, params)

		validator := types.Validator{SlashQueryProgressTracker: tc.progress}
		actualShouldQuery := s.App.StakeibcKeeper.ShouldCheckIfValidatorWasSlashed(s.Ctx, validator, tc.stakeAmount)
		s.Require().Equal(tc.expectedShouldQuery, actualShouldQuery, tc.name)
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
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		DelegationIcaAddress: delegationICAAddress,
		ConnectionId:         ibctesting.FirstConnectionID,
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
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZone, initalDeposit)
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

	// Remove connection ID and re-submit - should fail
	hostZoneWithoutConnectionId := hostZone
	hostZoneWithoutConnectionId.ConnectionId = ""
	err = s.App.StakeibcKeeper.DetokenizeLSMDeposit(s.Ctx, hostZoneWithoutConnectionId, initalDeposit)
	s.Require().ErrorContains(err, "unable to submit detokenization ICA")

	// Remove delegation account and re-submit - should also fail
	hostZoneWithoutDelegationAccount := hostZone
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
