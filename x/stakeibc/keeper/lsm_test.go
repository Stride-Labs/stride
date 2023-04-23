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
	validators := []*types.Validator{{Address: valAddress}}

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

func (s *KeeperTestSuite) TestRefundLSMToken() {
	liquidStakerAddress := s.TestAccs[0]
	despositAddress := types.NewHostZoneDepositAddress(HostChainId)

	// Fund the module account with the LSM token
	lsmTokenIBCDenom := "ibc/cosmosvalXXX"
	stakeAmount := sdk.NewInt(1000)
	lsmToken := sdk.NewCoin(lsmTokenIBCDenom, stakeAmount)
	s.FundAccount(despositAddress, lsmToken)

	// Setup the liquid stake object that provides the context for the refund
	liquidStake := types.LSMLiquidStake{
		Staker: liquidStakerAddress,
		HostZone: types.HostZone{
			DepositAddress: despositAddress.String(),
		},
		LSMIBCToken: lsmToken,
	}

	// Refund the token and check that it has been successfully transferred
	err := s.App.StakeibcKeeper.RefundLSMToken(s.Ctx, liquidStake)
	s.Require().NoError(err, "no error expected when refunding LSM token")

	stakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, liquidStakerAddress, lsmTokenIBCDenom)
	s.CompareCoins(lsmToken, stakerBalance, "staker should have received their LSM token back")

	moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, despositAddress, lsmTokenIBCDenom)
	s.True(moduleBalance.IsZero(), "module account should no longer have the LSM token")

	// Attempt to refund again, it should fail from an insufficient balance
	err = s.App.StakeibcKeeper.RefundLSMToken(s.Ctx, liquidStake)
	s.Require().ErrorContains(err, "insufficient funds")

	// Attempt to refund with an invalid host zone address, it should fail
	liquidStake.HostZone.DepositAddress = ""
	err = s.App.StakeibcKeeper.RefundLSMToken(s.Ctx, liquidStake)
	s.Require().ErrorContains(err, "host zone address is invalid")
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
					expectedStatus = recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS
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
