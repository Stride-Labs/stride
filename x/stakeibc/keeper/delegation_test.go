package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	proto "github.com/cosmos/gogoproto/proto"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	sdkmath "cosmossdk.io/math"

	epochstypes "github.com/Stride-Labs/stride/v23/x/epochs/types"
	epochtypes "github.com/Stride-Labs/stride/v23/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v23/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v23/x/records/types"
	"github.com/Stride-Labs/stride/v23/x/stakeibc/types"
)

type TestDepositRecords struct {
	emptyRecords          []recordstypes.DepositRecord
	recordsToBeTransfered []recordstypes.DepositRecord
	recordsToBeStaked     []recordstypes.DepositRecord
	recordsInCurrentEpoch []recordstypes.DepositRecord
	transferAmount        sdkmath.Int
	stakeAmount           sdkmath.Int
}

func (r *TestDepositRecords) GetAllRecords() []recordstypes.DepositRecord {
	allDepositRecords := []recordstypes.DepositRecord{}
	allDepositRecords = append(allDepositRecords, r.emptyRecords...)
	allDepositRecords = append(allDepositRecords, r.recordsToBeTransfered...)
	allDepositRecords = append(allDepositRecords, r.recordsToBeStaked...)
	allDepositRecords = append(allDepositRecords, r.recordsInCurrentEpoch...)
	return allDepositRecords
}

type Channel struct {
	PortID    string
	ChannelID string
}

type DepositRecordsTestCase struct {
	initialDepositRecords       TestDepositRecords
	initialModuleAccountBalance sdk.Coin
	hostZone                    types.HostZone
	hostZoneDepositAddress      sdk.AccAddress
	epochNumber                 uint64
	TransferChannel             Channel
	DelegationChannel           Channel
}

func (s *KeeperTestSuite) GetInitialDepositRecords(currentEpoch uint64) TestDepositRecords {
	priorEpoch := currentEpoch - 1
	emptyDepositRecords := []recordstypes.DepositRecord{
		{
			Id:                      1,
			Amount:                  sdkmath.ZeroInt(),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
		{
			Id:                      2,
			Amount:                  sdkmath.ZeroInt(),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
	}

	recordsToBeTransfered := []recordstypes.DepositRecord{
		{
			Id:                      3,
			Amount:                  sdkmath.NewInt(3000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
		{
			Id:                      4,
			Amount:                  sdkmath.NewInt(4000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
	}
	transferAmount := sdkmath.NewInt(3000 + 4000)

	recordsToBeStaked := []recordstypes.DepositRecord{
		{
			Id:                      5,
			Amount:                  sdkmath.NewInt(5000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
		{
			Id:                      6,
			Amount:                  sdkmath.NewInt(6000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber:      priorEpoch,
			DelegationTxsInProgress: 0,
		},
	}
	stakeAmount := sdkmath.NewInt(5000 + 6000)

	recordsInCurrentEpoch := []recordstypes.DepositRecord{
		{
			Id:                      7,
			Amount:                  sdkmath.NewInt(7000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber:      currentEpoch,
			DelegationTxsInProgress: 0,
		},
		{
			Id:                      8,
			Amount:                  sdkmath.NewInt(8000),
			Denom:                   Atom,
			HostZoneId:              HostChainId,
			Status:                  recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber:      currentEpoch,
			DelegationTxsInProgress: 0,
		},
	}

	return TestDepositRecords{
		emptyRecords:          emptyDepositRecords,
		recordsToBeTransfered: recordsToBeTransfered,
		recordsToBeStaked:     recordsToBeStaked,
		recordsInCurrentEpoch: recordsInCurrentEpoch,
		transferAmount:        transferAmount,
		stakeAmount:           stakeAmount,
	}
}

func (s *KeeperTestSuite) SetupDepositRecords() DepositRecordsTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	ibcDenomTrace := s.GetIBCDenomTrace(Atom) // we need a true IBC denom here
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, ibcDenomTrace)

	initialModuleAccountBalance := sdk.NewCoin(ibcDenomTrace.IBCDenom(), sdkmath.NewInt(15_000))
	s.FundAccount(depositAddress, initialModuleAccountBalance)

	validators := []*types.Validator{
		{
			Name:    "val1",
			Address: "gaia_VAL1",
			Weight:  1,
		},
		{
			Name:    "val2",
			Address: "gaia_VAL2",
			Weight:  4,
		},
	}

	hostZone := types.HostZone{
		ChainId:              HostChainId,
		DepositAddress:       depositAddress.String(),
		DelegationIcaAddress: delegationAddress,
		ConnectionId:         ibctesting.FirstConnectionID,
		TransferChannelId:    ibctesting.FirstChannelID,
		HostDenom:            Atom,
		IbcDenom:             ibcDenomTrace.IBCDenom(),
		Validators:           validators,
		MaxMessagesPerIcaTx:  10,
	}

	currentEpoch := uint64(2)
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	initialDepositRecords := s.GetInitialDepositRecords(currentEpoch)
	for _, depositRecord := range initialDepositRecords.GetAllRecords() {
		s.App.RecordsKeeper.AppendDepositRecord(s.Ctx, depositRecord)
	}

	return DepositRecordsTestCase{
		initialDepositRecords:       initialDepositRecords,
		initialModuleAccountBalance: initialModuleAccountBalance,
		hostZone:                    hostZone,
		hostZoneDepositAddress:      depositAddress,
		epochNumber:                 currentEpoch,
		TransferChannel: Channel{
			PortID:    ibctesting.TransferPort,
			ChannelID: ibctesting.FirstChannelID,
		},
		DelegationChannel: Channel{
			PortID:    delegationPortID,
			ChannelID: delegationChannelID,
		},
	}
}

// Helper function to check the state after transferring deposit records
// This assumes the last X transfers failed
func (s *KeeperTestSuite) CheckStateAfterTransferringDepositRecords(tc DepositRecordsTestCase, numTransfersFailed int) {
	// Get tx seq number before transfer to confirm that it gets incremented
	transferPortID := tc.TransferChannel.PortID
	transferChannelID := tc.TransferChannel.ChannelID
	startSequence := s.MustGetNextSequenceNumber(transferPortID, transferChannelID)

	// Transfer deposit records
	s.App.StakeibcKeeper.TransferExistingDepositsToHostZones(s.Ctx, tc.epochNumber, tc.initialDepositRecords.GetAllRecords())

	// Confirm tx sequence was incremented
	numTransferAttempts := len(tc.initialDepositRecords.recordsToBeTransfered)
	numSuccessfulTransfers := uint64(numTransferAttempts - numTransfersFailed)

	endSequence := s.MustGetNextSequenceNumber(transferPortID, transferChannelID)
	s.Require().Equal(startSequence+numSuccessfulTransfers, endSequence, "tx sequence number after transfer")

	// Confirm the callback data was stored for each transfer packet EXCLUDING the failed packets
	numCallbacks := uint64(len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)))
	s.Require().Equal(numSuccessfulTransfers, numCallbacks, "number of callbacks")

	recordsSuccessfullyTransferred := tc.initialDepositRecords.recordsToBeTransfered[:numSuccessfulTransfers]
	for i, depositRecord := range recordsSuccessfullyTransferred {
		// Confirm callback record
		callbackKey := icacallbackstypes.PacketID(transferPortID, transferChannelID, startSequence+uint64(i))
		callbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
		s.Require().True(found, "callback data was not found for callback key (%s)", callbackKey)
		s.Require().Equal("transfer", callbackData.CallbackId, "callback ID")

		// Confirm callback args
		callbackArgs, err := s.App.RecordsKeeper.UnmarshalTransferCallbackArgs(s.Ctx, callbackData.CallbackArgs)
		s.Require().NoError(err, "unmarshalling callback args error for callback key (%s)", callbackKey)
		s.Require().Equal(depositRecord.Id, callbackArgs.DepositRecordId, "deposit record ID in callback args (%s)", callbackKey)
	}

	// Confirm the module account balance decreased
	expectedTransferAmount := sdkmath.NewInt(0)
	for _, depositRecord := range recordsSuccessfullyTransferred {
		expectedTransferAmount = expectedTransferAmount.Add(depositRecord.Amount)
	}
	expectedModuleBalance := tc.initialModuleAccountBalance.SubAmount(expectedTransferAmount)
	actualModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.hostZoneDepositAddress, tc.hostZone.IbcDenom)
	s.CompareCoins(expectedModuleBalance, actualModuleBalance, "host module balance")

	// Confirm deposit records with 0 amount were removed
	expectedNumDepositRecords := len(tc.initialDepositRecords.GetAllRecords()) - len(tc.initialDepositRecords.emptyRecords)
	actualNumDepositRecords := len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx))
	s.Require().Equal(expectedNumDepositRecords, actualNumDepositRecords, "total deposit records")

	for _, emptyRecord := range tc.initialDepositRecords.emptyRecords {
		_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, emptyRecord.Id)
		s.Require().False(found, "empty deposit record (%d) should have been removed", emptyRecord.Id)
	}
}

func (s *KeeperTestSuite) TestTransferDepositRecords_Successful() {
	tc := s.SetupDepositRecords()

	numFailures := 0
	s.CheckStateAfterTransferringDepositRecords(tc, numFailures)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_HostZoneNotFound() {
	tc := s.SetupDepositRecords()
	// Replace first deposit record with a record that has a bad host zone
	recordsToBeTransfered := tc.initialDepositRecords.recordsToBeTransfered
	lastRecordIndex := len(recordsToBeTransfered) - 1

	badRecord := tc.initialDepositRecords.recordsToBeTransfered[lastRecordIndex]
	badRecord.HostZoneId = "fake_host_zone"
	tc.initialDepositRecords.recordsToBeTransfered[lastRecordIndex] = badRecord
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, badRecord)

	numFailed := 1
	s.CheckStateAfterTransferringDepositRecords(tc, numFailed)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_NoDelegationAccount() {
	tc := s.SetupDepositRecords()
	// Remove the delegation account from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeTransfered)
	s.CheckStateAfterTransferringDepositRecords(tc, numFailed)
}

// Helper function to check the state after staking deposit records
// This assumes the last X delegations failed
func (s *KeeperTestSuite) CheckStateAfterStakingDepositRecords(tc DepositRecordsTestCase, numDelegationsFailed int) {
	// Get tx seq number before delegation to confirm it incremented
	delegationPortID := tc.DelegationChannel.PortID
	delegationChannelID := tc.DelegationChannel.ChannelID
	startSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)

	// Stake deposit records
	s.App.StakeibcKeeper.StakeExistingDepositsOnHostZones(s.Ctx, tc.epochNumber, tc.initialDepositRecords.GetAllRecords())

	// Confirm tx sequence was incremented
	numDelegationAttempts := len(tc.initialDepositRecords.recordsToBeStaked)
	numSuccessfulDelegations := uint64(numDelegationAttempts - numDelegationsFailed)

	endSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)
	s.Require().Equal(startSequence+numSuccessfulDelegations, endSequence, "tx sequence number after delegation")

	// Confirm the callback data was stored for each delegation packet EXCLUDING the failed packets
	numCallbacks := uint64(len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)))
	s.Require().Equal(numSuccessfulDelegations, numCallbacks, "number of callback's stored")

	recordsSuccessfullyStaked := tc.initialDepositRecords.recordsToBeStaked[:numSuccessfulDelegations]
	for i, depositRecord := range recordsSuccessfullyStaked {
		// Confirm callback record
		callbackKey := icacallbackstypes.PacketID(delegationPortID, delegationChannelID, startSequence+uint64(i))
		callbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
		s.Require().True(found, "callback data was not found for callback key (%s)", callbackKey)
		s.Require().Equal("delegate", callbackData.CallbackId, "callback ID")

		// Confirm callback args
		callbackArgs := types.DelegateCallback{}
		err := proto.Unmarshal(callbackData.CallbackArgs, &callbackArgs)
		s.Require().NoError(err, "unmarshalling callback args error for callback key (%s)", callbackKey)
		s.Require().Equal(depositRecord.Id, callbackArgs.DepositRecordId, "deposit record ID in callback args (%s)", callbackKey)
		s.Require().Equal(tc.hostZone.ChainId, callbackArgs.HostZoneId, "host zone in callback args (%s)", callbackKey)

		// Confirm expected delegations
		val1 := tc.hostZone.Validators[0]
		val2 := tc.hostZone.Validators[1]
		totalWeight := val1.Weight + val2.Weight

		val1Delegation := depositRecord.Amount.Mul(sdkmath.NewIntFromUint64(val1.Weight)).Quo(sdkmath.NewIntFromUint64(totalWeight))
		val2Delegation := depositRecord.Amount.Mul(sdkmath.NewIntFromUint64(val2.Weight)).Quo(sdkmath.NewIntFromUint64(totalWeight))

		expectedDelegations := []*types.SplitDelegation{
			{Validator: val1.Address, Amount: val1Delegation},
			{Validator: val2.Address, Amount: val2Delegation},
		}

		s.Require().Equal(len(tc.hostZone.Validators), len(callbackArgs.SplitDelegations), "number of redelegations")
		for i := range expectedDelegations {
			s.Require().Equal(expectedDelegations[i], callbackArgs.SplitDelegations[i],
				"split delegations in callback args (%s), val (%s)", callbackKey, expectedDelegations[i].Validator)
		}

	}
}

func (s *KeeperTestSuite) TestStakeDepositRecords_Successful() {
	tc := s.SetupDepositRecords()

	numFailures := 0
	s.CheckStateAfterStakingDepositRecords(tc, numFailures)
}

func (s *KeeperTestSuite) TestStakeDepositRecords_HostZoneNotFound() {
	tc := s.SetupDepositRecords()
	// Replace first deposit record with a record that has a bad host zone
	recordsToBeStaked := tc.initialDepositRecords.recordsToBeStaked
	lastRecordIndex := len(recordsToBeStaked) - 1

	badRecord := tc.initialDepositRecords.recordsToBeStaked[lastRecordIndex]
	badRecord.HostZoneId = "fake_host_zone"
	tc.initialDepositRecords.recordsToBeStaked[lastRecordIndex] = badRecord
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, badRecord)

	numFailed := 1
	s.CheckStateAfterStakingDepositRecords(tc, numFailed)
}

func (s *KeeperTestSuite) TestStakeDepositRecords_NoDelegationAccount() {
	tc := s.SetupDepositRecords()
	// Remove the delegation account from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeStaked)
	s.CheckStateAfterStakingDepositRecords(tc, numFailed)
}

func (s *KeeperTestSuite) TestGetDelegationICAMessages() {
	delegationAddress := "cosmos_DELEGATION"

	testCases := []struct {
		name                string
		totalDelegated      sdkmath.Int
		validators          []*types.Validator
		expectedDelegations []types.SplitDelegation
		expectedError       string
	}{
		{
			name:           "one validator",
			totalDelegated: sdkmath.NewInt(50),
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
			},
			expectedDelegations: []types.SplitDelegation{
				{Validator: "val1", Amount: sdkmath.NewInt(50)},
			},
		},
		{
			name:           "two validators",
			totalDelegated: sdkmath.NewInt(100),
			validators: []*types.Validator{
				{Address: "val1", Weight: 1},
				{Address: "val2", Weight: 1},
			},
			expectedDelegations: []types.SplitDelegation{
				{Validator: "val1", Amount: sdkmath.NewInt(50)},
				{Validator: "val2", Amount: sdkmath.NewInt(50)},
			},
		},
		{
			name:           "three validators",
			totalDelegated: sdkmath.NewInt(100),
			validators: []*types.Validator{
				{Address: "val1", Weight: 25},
				{Address: "val2", Weight: 50},
				{Address: "val3", Weight: 25},
			},
			expectedDelegations: []types.SplitDelegation{
				{Validator: "val1", Amount: sdkmath.NewInt(25)},
				{Validator: "val2", Amount: sdkmath.NewInt(50)},
				{Validator: "val3", Amount: sdkmath.NewInt(25)},
			},
		},
		{
			name:           "zero weight validator",
			totalDelegated: sdkmath.NewInt(100),
			validators: []*types.Validator{
				{Address: "val1", Weight: 25},
				{Address: "val2", Weight: 0},
				{Address: "val3", Weight: 25},
			},
			expectedDelegations: []types.SplitDelegation{
				{Validator: "val1", Amount: sdkmath.NewInt(50)},
				{Validator: "val3", Amount: sdkmath.NewInt(50)},
			},
		},
		{
			name:           "zero weight validators",
			totalDelegated: sdkmath.NewInt(100),
			validators: []*types.Validator{
				{Address: "val1", Weight: 0},
				{Address: "val2", Weight: 0},
				{Address: "val3", Weight: 0},
			},
			expectedError: "No non-zero validators found",
		},
		{
			name:           "no validators",
			totalDelegated: sdkmath.NewInt(100),
			validators:     []*types.Validator{},
			expectedError:  "No non-zero validators found",
		},
		{
			name:           "zero total delegations",
			totalDelegated: sdkmath.NewInt(0),
			validators:     []*types.Validator{},
			expectedError:  "Cannot calculate target delegation if final amount is less than or equal to zero",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create the host zone and deposit record for the given test case
			hostZone := types.HostZone{
				ChainId:              HostChainId,
				HostDenom:            Atom,
				DelegationIcaAddress: delegationAddress,
				Validators:           tc.validators,
			}

			depositRecord := recordstypes.DepositRecord{
				Amount: tc.totalDelegated,
			}

			// Build the delegation ICA messages
			actualMessages, actualSplits, actualError := s.App.StakeibcKeeper.GetDelegationICAMessages(
				s.Ctx,
				hostZone,
				depositRecord,
			)

			// If this is an error test case, check the error message
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError, "error expected")
				return
			}

			// For the success case, check the error number of delegations
			s.Require().NoError(actualError, "no error expected when delegating %v", tc.expectedDelegations)
			s.Require().Len(actualMessages, len(tc.expectedDelegations), "number of undelegate messages")
			s.Require().Len(actualSplits, len(tc.expectedDelegations), "number of validator splits")

			// Check each delegation
			for i, expected := range tc.expectedDelegations {
				valAddress := expected.Validator
				actualMsg := actualMessages[i].(*stakingtypes.MsgDelegate)
				actualSplit := actualSplits[i]

				// Check the ICA message
				s.Require().Equal(valAddress, actualMsg.ValidatorAddress, "ica message validator")
				s.Require().Equal(delegationAddress, actualMsg.DelegatorAddress, "ica message delegator for %s", valAddress)
				s.Require().Equal(Atom, actualMsg.Amount.Denom, "ica message denom for %s", valAddress)
				s.Require().Equal(expected.Amount.Int64(), actualMsg.Amount.Amount.Int64(),
					"ica message amount for %s", valAddress)

				// Check the callback
				s.Require().Equal(expected.Validator, actualSplit.Validator, "callback validator for %s", valAddress)
				s.Require().Equal(expected.Amount.Int64(), actualSplit.Amount.Int64(), "callback amount %s", valAddress)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBatchSubmitDelegationICAMessages() {
	// The test will submit ICA's across 10 validators, in batches of 3
	// There should be 4 ICA's submitted
	batchSize := 3
	numValidators := 10
	expectedNumberOfIcas := 4
	depositRecord := recordstypes.DepositRecord{}

	// Create the delegation ICA channel
	delegationAccountOwner := types.FormatHostZoneICAOwner(HostChainId, types.ICAAccountType_DELEGATION)
	delegationChannelID, delegationPortID := s.CreateICAChannel(delegationAccountOwner)

	// Create a host zone
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
	}

	// Build the ICA messages and callback for each validator
	var validators []*types.Validator
	var undelegateMsgs []proto.Message
	var delegations []*types.SplitDelegation
	for i := 0; i < numValidators; i++ {
		validatorAddress := fmt.Sprintf("val%d", i)
		validators = append(validators, &types.Validator{Address: validatorAddress})

		undelegateMsgs = append(undelegateMsgs, &stakingtypes.MsgDelegate{
			DelegatorAddress: hostZone.DelegationIcaAddress,
			ValidatorAddress: validatorAddress,
			Amount:           sdk.NewCoin(hostZone.HostDenom, sdkmath.NewInt(100)),
		})

		delegations = append(delegations, &types.SplitDelegation{
			Validator: validatorAddress,
			Amount:    sdkmath.NewInt(100),
		})
	}

	// Store the validators on the host zone
	hostZone.Validators = validators
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Mock the epoch tracker to timeout 90% through the epoch
	strideEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochstypes.STRIDE_EPOCH,
		Duration:           10_000_000_000,                                                // 10 second epochs
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeout
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get tx seq number before the ICA was submitted to check whether an ICA was submitted
	startSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)

	// Submit the delegations
	numTxsSubmitted, err := s.App.StakeibcKeeper.BatchSubmitDelegationICAMessages(
		s.Ctx,
		hostZone,
		depositRecord,
		undelegateMsgs,
		delegations,
		batchSize,
	)
	s.Require().NoError(err, "no error expected when submitting batches")
	s.Require().Equal(numTxsSubmitted, uint64(expectedNumberOfIcas), "returned number of txs submitted")

	// Confirm the sequence number iterated by the expected number of ICAs
	endSequence := s.MustGetNextSequenceNumber(delegationPortID, delegationChannelID)
	s.Require().Equal(startSequence+uint64(expectedNumberOfIcas), endSequence, "expected number of ICA submissions")

	// Confirm the number of callback data's matches the expected number of ICAs
	callbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Equal(expectedNumberOfIcas, len(callbackData), "number of callback datas")

	// Remove the connection ID from the host zone and try again, it should fail
	invalidHostZone := hostZone
	invalidHostZone.ConnectionId = ""
	_, err = s.App.StakeibcKeeper.BatchSubmitDelegationICAMessages(
		s.Ctx,
		invalidHostZone,
		depositRecord,
		undelegateMsgs,
		delegations,
		batchSize,
	)
	s.Require().ErrorContains(err, "failed to submit delegation ICAs")
}
