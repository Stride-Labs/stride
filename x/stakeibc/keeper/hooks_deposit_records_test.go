package keeper_test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type testDepositRecords struct {
	emptyDepositRecords   []recordstypes.DepositRecord
	recordsToBeTransfered []recordstypes.DepositRecord
	recordsToBeStaked     []recordstypes.DepositRecord
	recordsInCurrentEpoch []recordstypes.DepositRecord
	allDepositRecords     []recordstypes.DepositRecord
}

func (s *KeeperTestSuite) GetInitialDepositRecords(currentEpoch uint64) testDepositRecords {
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

	recordsInCurrentEpoch := []recordstypes.DepositRecord{
		{
			Id:                 5,
			Amount:             5000,
			Denom:              Atom,
			HostZoneId:         HostChainId,
			Status:             recordstypes.DepositRecord_STAKE,
			DepositEpochNumber: currentEpoch,
		},
		{
			Id:                 6,
			Amount:             6000,
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

	return testDepositRecords{
		emptyDepositRecords:   emptyDepositRecords,
		recordsToBeTransfered: recordsToBeTransfered,
		recordsToBeStaked:     recordsToBeStaked,
		recordsInCurrentEpoch: recordsInCurrentEpoch,
		allDepositRecords:     allDepositRecords,
	}
}

func (s *KeeperTestSuite) SetupHooksDepositRecords() {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	hostZone := stakeibctypes.HostZone{
		ChainId: HostChainId,
		Address: delegationAddress,
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
}

func (s *KeeperTestSuite) TestTransferDepositRecords_Successful() {

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
