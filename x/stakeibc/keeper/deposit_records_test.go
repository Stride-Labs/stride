package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type TestDepositRecords struct {
	emptyRecords          []recordstypes.DepositRecord
	recordsToBeTransfered []recordstypes.DepositRecord
	recordsToBeStaked     []recordstypes.DepositRecord
	recordsInCurrentEpoch []recordstypes.DepositRecord
	transferAmount        sdk.Int
	stakeAmount           sdk.Int
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
	hostZone                    stakeibctypes.HostZone
	hostModuleAddress           sdk.AccAddress
	epochNumber                 uint64
	TransferChannel             Channel
	DelegationChannel           Channel
}

func (s *KeeperTestSuite) GetInitialDepositRecords(currentEpoch uint64) TestDepositRecords {
	priorEpoch := currentEpoch - 1
	emptyDepositRecords := []recordstypes.DepositRecord{
		{
			Id:                 1,
			Amount:             sdk.ZeroInt(),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 2,
			Amount:             sdk.ZeroInt(),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
	}

	recordsToBeTransfered := []recordstypes.DepositRecord{
		{
			Id:                 3,
			Amount:             sdk.NewInt(3000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 4,
			Amount:             sdk.NewInt(4000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
	}
	transferAmount := sdk.NewInt(3000 + 4000)

	recordsToBeStaked := []recordstypes.DepositRecord{
		{
			Id:                 5,
			Amount:             sdk.NewInt(5000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 6,
			Amount:             sdk.NewInt(6000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber: priorEpoch,
		},
	}
	stakeAmount := sdk.NewInt(5000 + 6000)

	recordsInCurrentEpoch := []recordstypes.DepositRecord{
		{
			Id:                 7,
			Amount:             sdk.NewInt(7000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber: currentEpoch,
		},
		{
			Id:                 8,
			Amount:            sdk.NewInt(8000),
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
			DepositEpochNumber: currentEpoch,
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
	delegationChannelID := s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	ibcDenomTrace := s.GetIBCDenomTrace(Atom) // we need a true IBC denom here
	hostModuleAddress := stakeibctypes.NewZoneAddress(HostChainId)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, ibcDenomTrace)

	initialModuleAccountBalance := sdk.NewCoin(ibcDenomTrace.IBCDenom(), sdk.NewInt(15_000))
	s.FundAccount(hostModuleAddress, initialModuleAccountBalance)

	validators := []*stakeibctypes.Validator{
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

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		Address:           hostModuleAddress.String(),
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
		ConnectionId:      ibctesting.FirstConnectionID,
		TransferChannelId: ibctesting.FirstChannelID,
		HostDenom:         Atom,
		IbcDenom:          ibcDenomTrace.IBCDenom(),
		Validators:        validators,
	}

	currentEpoch := uint64(2)
	strideEpochTracker := stakeibctypes.EpochTracker{
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
		hostModuleAddress:           hostModuleAddress,
		epochNumber:                 currentEpoch,
		TransferChannel: Channel{
			PortID:    ibctesting.TransferPort,
			ChannelID: ibctesting.FirstChannelID,
		},
		DelegationChannel: Channel{
			PortID:    icatypes.PortPrefix + delegationAccountOwner,
			ChannelID: delegationChannelID,
		},
	}
}

func (s *KeeperTestSuite) TestCreateDepositRecordsForEpoch_Successful() {
	// Set host zones
	hostZones := []stakeibctypes.HostZone{
		{
			ChainId:   "HOST1",
			HostDenom: "denom1",
		},
		{
			ChainId:   "HOST2",
			HostDenom: "denom2",
		},
		{
			ChainId:   "HOST3",
			HostDenom: "denom3",
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Create depoist records for two epochs
	s.App.StakeibcKeeper.CreateDepositRecordsForEpoch(s.Ctx, 1)
	s.App.StakeibcKeeper.CreateDepositRecordsForEpoch(s.Ctx, 2)

	expectedDepositRecords := []recordstypes.DepositRecord{
		// Epoch 1
		{
			Id:                 0,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom1",
			HostZoneId:         "HOST1",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		{
			Id:                 1,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom2",
			HostZoneId:         "HOST2",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		{
			Id:                 2,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom3",
			HostZoneId:         "HOST3",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		// Epoch 2
		{
			Id:                 3,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom1",
			HostZoneId:         "HOST1",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
		{
			Id:                 4,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom2",
			HostZoneId:         "HOST2",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
		{
			Id:                 5,
			Amount:             sdk.ZeroInt(),
			Denom:              "denom3",
			HostZoneId:         "HOST3",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
	}

	// Confirm deposit records
	actualDepositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Equal(len(expectedDepositRecords), len(actualDepositRecords), "number of deposit records")
	s.Require().Equal(expectedDepositRecords, actualDepositRecords, "deposit records")
}

// Helper function to check the state after transferring deposit records
// This assumes the last X transfers failed
func (s *KeeperTestSuite) CheckStateAfterTransferringDepositRecords(tc DepositRecordsTestCase, numTransfersFailed int) {
	// Get tx seq number before transfer to confirm that it gets incremented
	transferPortID := tc.TransferChannel.PortID
	transferChannelID := tc.TransferChannel.ChannelID
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, transferPortID, transferChannelID)
	s.Require().True(found, "sequence number not found before transfer")

	// Transfer deposit records
	s.App.StakeibcKeeper.TransferExistingDepositsToHostZones(s.Ctx, tc.epochNumber, tc.initialDepositRecords.GetAllRecords())

	// Confirm tx sequence was incremented
	numTransferAttempts := len(tc.initialDepositRecords.recordsToBeTransfered)
	numSuccessfulTransfers := uint64(numTransferAttempts - numTransfersFailed)

	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, transferPortID, transferChannelID)
	s.Require().True(found, "sequence number not found after transfer")
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
	expectedTransferAmount := sdk.NewInt(0)
	for _, depositRecord := range recordsSuccessfullyTransferred {
		expectedTransferAmount = expectedTransferAmount.Add(depositRecord.Amount)
	}
	expectedModuleBalance := tc.initialModuleAccountBalance.SubAmount(expectedTransferAmount)
	actualModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.hostModuleAddress, tc.hostZone.IbcDenom)
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
	badHostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeTransfered)
	s.CheckStateAfterTransferringDepositRecords(tc, numFailed)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_NoDelegationAddress() {
	tc := s.SetupDepositRecords()
	// Remove the delegation address from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationAccount.Address = ""
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
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found before delegation")

	// Stake deposit records
	s.App.StakeibcKeeper.StakeExistingDepositsOnHostZones(s.Ctx, tc.epochNumber, tc.initialDepositRecords.GetAllRecords())

	// Confirm tx sequence was incremented
	numDelegationAttempts := len(tc.initialDepositRecords.recordsToBeStaked)
	numSuccessfulDelegations := uint64(numDelegationAttempts - numDelegationsFailed)

	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortID, delegationChannelID)
	s.Require().True(found, "sequence number not found after delegation")
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
		callbackArgs, err := s.App.StakeibcKeeper.UnmarshalDelegateCallbackArgs(s.Ctx, callbackData.CallbackArgs)
		s.Require().NoError(err, "unmarshalling callback args error for callback key (%s)", callbackKey)
		s.Require().Equal(depositRecord.Id, callbackArgs.DepositRecordId, "deposit record ID in callback args (%s)", callbackKey)
		s.Require().Equal(tc.hostZone.ChainId, callbackArgs.HostZoneId, "host zone in callback args (%s)", callbackKey)

		// Confirm expected delegations
		val1 := tc.hostZone.Validators[0]
		val2 := tc.hostZone.Validators[1]
		totalWeight := val1.Weight + val2.Weight

		val1Delegation := depositRecord.Amount.Mul(sdk.NewIntFromUint64(val1.Weight)).Quo(sdk.NewIntFromUint64(totalWeight))
		val2Delegation := depositRecord.Amount.Mul(sdk.NewIntFromUint64(val2.Weight)).Quo(sdk.NewIntFromUint64(totalWeight))

		expectedDelegations := []*stakeibctypes.SplitDelegation{
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

func (s *KeeperTestSuite) TestStakeDepositRecords_SuccessfulCapped() {
	tc := s.SetupDepositRecords()

	// Set the cap on the number of deposit records processed to 1
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.MaxStakeIcaCallsPerEpoch = 1
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// The cap should cause the last record to not get processed
	numFailures := 1
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
	badHostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeStaked)
	s.CheckStateAfterStakingDepositRecords(tc, numFailed)
}

func (s *KeeperTestSuite) TestStakeDepositRecords_NoDelegationAddress() {
	tc := s.SetupDepositRecords()
	// Remove the delegation address from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationAccount.Address = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeStaked)
	s.CheckStateAfterStakingDepositRecords(tc, numFailed)
}
