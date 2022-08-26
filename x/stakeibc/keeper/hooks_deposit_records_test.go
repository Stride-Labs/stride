package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

// QUESTION: Does anyone have a better suggestion for the file organization of these two tests?
// They share the same setup so I originally thought it made sense to put them in the same file,
// but it's growing pretty large
// Also not married to the filename if anyone has a better suggestion!
// Another option:
//   |-- hooks_deposit_records_test.go          // setup one
//   |-- hooks_transfer_deposit_records_test.go // transfer tests
//   |-- hooks_stake_deposit_records_test.go    // stake tests
type TestDepositRecords struct {
	emptyDepositRecords   []recordstypes.DepositRecord
	recordsToBeTransfered []recordstypes.DepositRecord
	recordsToBeStaked     []recordstypes.DepositRecord
	recordsInCurrentEpoch []recordstypes.DepositRecord
	allDepositRecords     []recordstypes.DepositRecord
	transferAmount        sdk.Int
	stakeAmount           sdk.Int
}

type HooksDepositRecordsTestCase struct {
	initialDepositRecords       TestDepositRecords
	initialModuleAccountBalance sdk.Coin
	hostZone                    stakeibctypes.HostZone
	hostModuleAddress           sdk.AccAddress
	epochNumber                 uint64
}

func (s *KeeperTestSuite) GetInitialDepositRecords(currentEpoch uint64) TestDepositRecords {
	priorEpoch := currentEpoch - 1
	emptyDepositRecords := []recordstypes.DepositRecord{
		{
			Id:                 1,
			Amount:             0,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 2,
			Amount:             0,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: priorEpoch,
		},
	}

	recordsToBeTransfered := []recordstypes.DepositRecord{
		{
			Id:                 3,
			Amount:             3000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 4,
			Amount:             4000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: priorEpoch,
		},
	}
	transferAmount := sdk.NewInt(3000 + 4000)

	recordsToBeStaked := []recordstypes.DepositRecord{
		{
			Id:                 5,
			Amount:             5000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			DepositEpochNumber: priorEpoch,
		},
		{
			Id:                 6,
			Amount:             6000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			DepositEpochNumber: priorEpoch,
		},
	}
	stakeAmount := sdk.NewInt(5000 + 6000)

	recordsInCurrentEpoch := []recordstypes.DepositRecord{
		{
			Id:                 7,
			Amount:             7000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			DepositEpochNumber: currentEpoch,
		},
		{
			Id:                 8,
			Amount:             8000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			DepositEpochNumber: currentEpoch,
		},
	}

	allDepositRecords := []recordstypes.DepositRecord{}
	allDepositRecords = append(allDepositRecords, emptyDepositRecords...)
	allDepositRecords = append(allDepositRecords, recordsToBeTransfered...)
	allDepositRecords = append(allDepositRecords, recordsToBeStaked...)
	allDepositRecords = append(allDepositRecords, recordsInCurrentEpoch...)

	return TestDepositRecords{
		emptyDepositRecords:   emptyDepositRecords,
		recordsToBeTransfered: recordsToBeTransfered,
		recordsToBeStaked:     recordsToBeStaked,
		recordsInCurrentEpoch: recordsInCurrentEpoch,
		allDepositRecords:     allDepositRecords,
		transferAmount:        transferAmount,
		stakeAmount:           stakeAmount,
	}
}

func (s *KeeperTestSuite) SetupHooksDepositRecords() HooksDepositRecordsTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	ibcDenomTrace := s.GetIBCDenom(Atom) // we need a true IBC denom here
	hostModuleAddress := stakeibctypes.NewZoneAddress(HostChainId)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx(), ibcDenomTrace)

	initialModuleAccountBalance := sdk.NewCoin(ibcDenomTrace.IBCDenom(), sdk.NewInt(15_000))
	s.FundAccount(hostModuleAddress, initialModuleAccountBalance)

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		Address:           hostModuleAddress.String(),
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
		ConnectionId:      ibctesting.FirstConnectionID,
		TransferChannelId: ibctesting.FirstChannelID,
		IBCDenom:          ibcDenomTrace.IBCDenom(),
	}

	currentEpoch := uint64(2)
	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     currentEpoch,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)

	initialDepositRecords := s.GetInitialDepositRecords(currentEpoch)
	for _, depositRecord := range initialDepositRecords.allDepositRecords {
		s.App.RecordsKeeper.AppendDepositRecord(s.Ctx(), depositRecord)
	}

	return HooksDepositRecordsTestCase{
		initialDepositRecords:       initialDepositRecords,
		initialModuleAccountBalance: initialModuleAccountBalance,
		hostZone:                    hostZone,
		hostModuleAddress:           hostModuleAddress,
		epochNumber:                 currentEpoch,
	}
}

func (s *KeeperTestSuite) TestTransferDepositRecords_Successful() {
	tc := s.SetupHooksDepositRecords()

	// Get tx seq number before transfer to confirm it incremented
	transferPort := ibctesting.TransferPort
	transferChannel := ibctesting.FirstChannelID
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	s.Require().True(found, "get next sequence number not found")

	// Transfer deposit records
	s.App.StakeibcKeeper.TransferExistingDepositsToHostZones(s.Ctx(), tc.epochNumber, tc.initialDepositRecords.allDepositRecords)

	// Confirm tx sequence was incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	numTransfers := uint64(len(tc.initialDepositRecords.recordsToBeTransfered))
	s.Require().Equal(startSequence+numTransfers, endSequence, "tx sequence number after transfer")

	// Confirm the callback data was stored for each transfer packet
	numCallbacks := uint64(len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx())))
	s.Require().Equal(numTransfers, numCallbacks, "number of callback data's stored")
	for i := range tc.initialDepositRecords.recordsToBeTransfered {
		callbackKey := icacallbackstypes.PacketID(transferPort, transferChannel, startSequence+uint64(i))
		actualCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx(), callbackKey)
		_ = actualCallbackData
		s.Require().True(found, "callback data was not found for callback ID (%s)", callbackKey)
	}

	// Confirm the module account balance decreased
	expectedModuleBalance := tc.initialModuleAccountBalance.SubAmount(tc.initialDepositRecords.transferAmount)
	actualModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx(), tc.hostModuleAddress, tc.hostZone.IBCDenom)
	s.CompareCoins(expectedModuleBalance, actualModuleBalance, "host module balance")

	// Confirm deposit records with 0 amount were removed
	expectedNumDepositRecords := len(tc.initialDepositRecords.allDepositRecords) - len(tc.initialDepositRecords.emptyDepositRecords)
	actualNumDepositRecords := len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx()))
	s.Require().Equal(expectedNumDepositRecords, actualNumDepositRecords)

	for _, emptyRecord := range tc.initialDepositRecords.emptyDepositRecords {
		_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), emptyRecord.Id)
		s.Require().False(found, "empty deposit record (%d) should be removed", emptyRecord.Id)
	}
}

func (s *KeeperTestSuite) TestTransferDepositRecords_SuccessfulTransferMsg() {

}

// Helper function for the TransferDepositRecrods error cases
// This confirms everything that's done in the success case,
// but with the assumption that the the first X deposit records failed
func (s *KeeperTestSuite) CheckStateWithFailedTransfers(tc HooksDepositRecordsTestCase, numberTransfersFailed int) {
	// Get tx seq number before transfer to confirm it incremented
	transferPort := ibctesting.TransferPort
	transferChannel := ibctesting.FirstChannelID
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	s.Require().True(found, "get next sequence number not found")

	// Transfer deposit records
	s.App.StakeibcKeeper.TransferExistingDepositsToHostZones(s.Ctx(), tc.epochNumber, tc.initialDepositRecords.allDepositRecords)

	// Confirm tx sequence was incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	numTransfers := uint64(len(tc.initialDepositRecords.recordsToBeTransfered) - numberTransfersFailed) // exclude failures
	s.Require().Equal(startSequence+numTransfers, endSequence, "tx sequence number after transfer")

	// Confirm the callback data was stored for each transfer packet EXCLUDING the failed packets
	numCallbacks := uint64(len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx())))
	s.Require().Equal(numTransfers, numCallbacks, "number of callback data's stored")
	for i := range tc.initialDepositRecords.recordsToBeTransfered[numberTransfersFailed:] {
		callbackKey := icacallbackstypes.PacketID(transferPort, transferChannel, startSequence+uint64(i))
		actualCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx(), callbackKey)
		_ = actualCallbackData
		s.Require().True(found, "callback data was not found for callback ID (%s)", callbackKey)
	}

	// Confirm the module account balance decreased
	expectedTransferAmount := sdk.NewInt(0)
	for _, depositRecord := range tc.initialDepositRecords.recordsToBeTransfered[numberTransfersFailed:] {
		expectedTransferAmount = expectedTransferAmount.AddRaw(depositRecord.Amount)
	}
	expectedModuleBalance := tc.initialModuleAccountBalance.SubAmount(expectedTransferAmount)
	actualModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx(), tc.hostModuleAddress, tc.hostZone.IBCDenom)
	s.CompareCoins(expectedModuleBalance, actualModuleBalance, "host module balance")

	// Confirm deposit records with 0 amount were removed
	expectedNumDepositRecords := len(tc.initialDepositRecords.allDepositRecords) - len(tc.initialDepositRecords.emptyDepositRecords)
	actualNumDepositRecords := len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx()))
	s.Require().Equal(expectedNumDepositRecords, actualNumDepositRecords)

	for _, emptyRecord := range tc.initialDepositRecords.emptyDepositRecords {
		_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), emptyRecord.Id)
		s.Require().False(found, "empty deposit record (%d) should be removed", emptyRecord.Id)
	}
}

func (s *KeeperTestSuite) TestTransferDepositRecords_HostZoneNotFound() {
	tc := s.SetupHooksDepositRecords()
	// Replace first deposit record with a record that has a bad host zone
	badRecord := tc.initialDepositRecords.recordsToBeTransfered[0]
	badRecord.HostZoneId = "fake_host_zone"
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx(), badRecord)

	tc.initialDepositRecords.recordsToBeTransfered[0] = badRecord
	for i, depositRecord := range tc.initialDepositRecords.allDepositRecords {
		if depositRecord.Id == badRecord.Id {
			tc.initialDepositRecords.allDepositRecords[i] = badRecord
			break
		}
	}

	numFailed := 1
	s.CheckStateWithFailedTransfers(tc, numFailed)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_NoDelegationAccount() {
	tc := s.SetupHooksDepositRecords()
	// Remove the delegation account from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeTransfered)
	s.CheckStateWithFailedTransfers(tc, numFailed)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_NoDelegationAddress() {
	tc := s.SetupHooksDepositRecords()
	// Remove the delegation address from the host zone
	badHostZone := tc.hostZone
	badHostZone.DelegationAccount.Address = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeTransfered)
	s.CheckStateWithFailedTransfers(tc, numFailed)
}

func (s *KeeperTestSuite) TestTransferDepositRecords_BlockHeightNotFound() {
	tc := s.SetupHooksDepositRecords()
	// Remove the connection ID from the host zone
	badHostZone := tc.hostZone
	badHostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), badHostZone)

	numFailed := len(tc.initialDepositRecords.recordsToBeTransfered)
	s.CheckStateWithFailedTransfers(tc, numFailed)
}

func (s *KeeperTestSuite) TestStakeDepositRecords_Successful() {

}

func (s *KeeperTestSuite) TestStakeDepositRecords_SuccessfulCapped() {

}

func (s *KeeperTestSuite) TestStakeDepositRecords_SuccessfulDelegationMsg() {

}

func (s *KeeperTestSuite) TestStakeDepositRecords_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestStakeDepositRecords_NoDelegationAddress() {

}
