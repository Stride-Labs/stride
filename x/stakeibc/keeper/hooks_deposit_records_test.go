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
	initialDepositRecords := tc.initialDepositRecords

	// Get tx seq number before transfer to confirm it incremented
	transferPort := ibctesting.TransferPort
	transferChannel := ibctesting.FirstChannelID
	startNextSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	s.Require().True(found, "get next sequence number not found")

	// Transfer deposit records
	s.App.StakeibcKeeper.TransferExistingDepositsToHostZones(s.Ctx(), tc.epochNumber, initialDepositRecords.allDepositRecords)

	// Confirm tx sequence was incremented
	endNextSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx(), transferPort, transferChannel)
	numTransfers := uint64(len(tc.initialDepositRecords.recordsToBeTransfered))
	s.Require().Equal(endNextSequence, startNextSequence+numTransfers, "tx sequence number")

	// Confirm the callback data was stored for each transfer packet
	for i := range tc.initialDepositRecords.recordsToBeTransfered {
		callbackKey := icacallbackstypes.PacketID(transferPort, transferChannel, startNextSequence+uint64(i))
		actualCallbackData, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx(), callbackKey)
		_ = actualCallbackData
		s.Require().True(found, "callback data was not found for callback ID (%s)", callbackKey)
	}

	// Confirm deposit records with 0 amount were removed
	expectedNumDepositRecords := len(initialDepositRecords.allDepositRecords) - len(initialDepositRecords.emptyDepositRecords)
	actualNumDepositRecords := len(s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx()))
	s.Require().Equal(expectedNumDepositRecords, actualNumDepositRecords)

	for _, emptyRecord := range initialDepositRecords.emptyDepositRecords {
		_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx(), emptyRecord.Id)
		s.Require().False(found, "empty deposit record (%d) should be removed", emptyRecord.Id)
	}

	// Confirm the module account balance decreased
	expectedModuleBalance := tc.initialModuleAccountBalance.SubAmount(tc.initialDepositRecords.transferAmount)
	actualModuleBalance := s.App.BankKeeper.GetBalance(s.Ctx(), tc.hostModuleAddress, tc.hostZone.IBCDenom)
	s.CompareCoins(expectedModuleBalance, actualModuleBalance, "host module balance")
}

func (s *KeeperTestSuite) TestTransferDepositRecords_SuccessfulTransferMsg() {

}

func (s *KeeperTestSuite) TestTransferDepositRecords_HostZoneNotFound() {

}

func (s *KeeperTestSuite) TestTransferDepositRecords_NoDelegationAddress() {

}

func (s *KeeperTestSuite) TestTransferDepositRecords_BlockHeightNotFound() {

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
