package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/x/records/types"
)

// Helper function to create a new host zone unbonding record, filling in the sdkmath.Int's
// so that they can be compared
func newHostZoneUnbonding(chainId string, status types.HostZoneUnbonding_Status) types.HostZoneUnbonding {
	return types.HostZoneUnbonding{
		HostZoneId:            chainId,
		Status:                status,
		StTokenAmount:         sdkmath.ZeroInt(),
		NativeTokenAmount:     sdkmath.ZeroInt(),
		NativeTokensToUnbond:  sdkmath.ZeroInt(),
		StTokensToBurn:        sdkmath.ZeroInt(),
		ClaimableNativeTokens: sdkmath.ZeroInt(),
	}
}

func (s *KeeperTestSuite) createNEpochUnbondingRecord(n int) ([]types.EpochUnbondingRecord, map[string]types.HostZoneUnbonding) {
	hostZoneUnbondingsList := []types.HostZoneUnbonding{
		newHostZoneUnbonding("host-A", types.HostZoneUnbonding_UNBONDING_QUEUE),
		newHostZoneUnbonding("host-B", types.HostZoneUnbonding_UNBONDING_QUEUE),
		newHostZoneUnbonding("host-C", types.HostZoneUnbonding_UNBONDING_QUEUE),
	}
	hostZoneUnbondingsMap := make(map[string]types.HostZoneUnbonding)
	for _, hostZoneUnbonding := range hostZoneUnbondingsList {
		hostZoneUnbondingsMap[hostZoneUnbonding.HostZoneId] = hostZoneUnbonding
	}

	epochUnbondingRecords := make([]types.EpochUnbondingRecord, n)
	for epochNumber, epochUnbondingRecord := range epochUnbondingRecords {
		epochUnbondingRecord.EpochNumber = uint64(epochNumber)

		unbondingsCopy := make([]*types.HostZoneUnbonding, 3)
		for i := range unbondingsCopy {
			hostZoneUnbonding := hostZoneUnbondingsList[i]
			epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &hostZoneUnbonding)
		}

		epochUnbondingRecords[epochNumber] = epochUnbondingRecord
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}
	return epochUnbondingRecords, hostZoneUnbondingsMap
}

func (s *KeeperTestSuite) TestEpochUnbondingRecordGet() {
	items, _ := s.createNEpochUnbondingRecord(10)
	for _, item := range items {
		got, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, item.EpochNumber)
		s.Require().True(found)
		s.Require().Equal(
			&item,
			&got,
		)
	}
}

func (s *KeeperTestSuite) TestEpochUnbondingRecordRemove() {
	items, _ := s.createNEpochUnbondingRecord(10)
	for _, item := range items {
		s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, item.EpochNumber)
		_, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, item.EpochNumber)
		s.Require().False(found)
	}
}

func (s *KeeperTestSuite) TestEpochUnbondingRecordGetAll() {
	items, _ := s.createNEpochUnbondingRecord(10)
	s.Require().ElementsMatch(
		items,
		s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx),
	)
}

func (s *KeeperTestSuite) TestGetAllPreviousEpochUnbondingRecords() {
	items, _ := s.createNEpochUnbondingRecord(10)
	currentEpoch := uint64(8)
	fetchedItems := items[:currentEpoch]
	s.Require().ElementsMatch(
		fetchedItems,
		s.App.RecordsKeeper.GetAllPreviousEpochUnbondingRecords(s.Ctx, currentEpoch),
	)
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingByChainId() {
	_, hostZoneUnbondings := s.createNEpochUnbondingRecord(10)

	expectedHostZoneUnbonding := hostZoneUnbondings["host-B"]
	actualHostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, 1, "host-B")

	s.Require().True(found)
	s.Require().Equal(
		*actualHostZoneUnbonding,
		expectedHostZoneUnbonding,
	)
}

func (s *KeeperTestSuite) TestAddHostZoneToEpochUnbondingRecord() {
	epochUnbondingRecords, _ := s.createNEpochUnbondingRecord(3)

	epochNumber := uint64(0)
	initialEpochUnbondingRecord := epochUnbondingRecords[int(epochNumber)]

	// Update host zone unbonding for host-C
	updatedHostZoneUnbonding := newHostZoneUnbonding("host-C", types.HostZoneUnbonding_UNBONDING_IN_PROGRESS)

	expectedEpochUnbondingRecord := initialEpochUnbondingRecord
	expectedEpochUnbondingRecord.HostZoneUnbondings[2] = &updatedHostZoneUnbonding

	updatedEpochUnbonding, err := s.App.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(s.Ctx, epochNumber, "host-C", updatedHostZoneUnbonding)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, updatedEpochUnbonding)
	s.Require().NoError(err, "no error expected when updating host-C")
	for i := 0; i < len(expectedEpochUnbondingRecord.HostZoneUnbondings); i++ {
		expectedHostZoneUnbonding := *expectedEpochUnbondingRecord.HostZoneUnbondings[i]
		actualHostZoneUnbonding := *updatedEpochUnbonding.HostZoneUnbondings[i]
		s.Require().Equal(expectedHostZoneUnbonding, actualHostZoneUnbonding, "HZU %d after host-C update", i)
	}

	// Add new host zone to initial epoch unbonding records
	newHostZoneUnbonding := newHostZoneUnbonding("host-D", types.HostZoneUnbonding_UNBONDING_QUEUE)
	expectedEpochUnbondingRecord.HostZoneUnbondings = append(expectedEpochUnbondingRecord.HostZoneUnbondings, &newHostZoneUnbonding)

	updatedEpochUnbonding, err = s.App.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(s.Ctx, epochNumber, "host-D", newHostZoneUnbonding)
	s.Require().NoError(err, "no error expected when adding host-D")
	s.Require().Equal(expectedEpochUnbondingRecord, updatedEpochUnbonding, "EUR after host-D addition")
}

func (s *KeeperTestSuite) TestSetHostZoneUnbondingStatus() {
	initialEpochUnbondingRecords, _ := s.createNEpochUnbondingRecord(4)

	epochsToUpdate := []uint64{1, 3}
	hostIdToUpdate := "host-B"
	newStatus := types.HostZoneUnbonding_UNBONDING_IN_PROGRESS

	expectedEpochUnbondingRecords := initialEpochUnbondingRecords
	for _, epochUnbondingRecord := range expectedEpochUnbondingRecords {
		for _, epochNumberToUpdate := range epochsToUpdate {
			if epochUnbondingRecord.EpochNumber == epochNumberToUpdate {
				for i, hostUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
					if hostUnbonding.HostZoneId == hostIdToUpdate {
						updatedHostZoneUnbonding := hostUnbonding
						updatedHostZoneUnbonding.Status = newStatus
						epochUnbondingRecord.HostZoneUnbondings[i] = updatedHostZoneUnbonding
					}
				}
			}
		}
	}

	err := s.App.RecordsKeeper.SetHostZoneUnbondingStatus(s.Ctx, hostIdToUpdate, epochsToUpdate, newStatus)
	s.Require().Nil(err)

	actualEpochUnbondingRecord := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	s.Require().ElementsMatch(
		expectedEpochUnbondingRecords,
		actualEpochUnbondingRecord,
	)
}

func (s *KeeperTestSuite) TestSetHostZoneUnbonding() {
	initialAmount := sdkmath.NewInt(10)
	updatedAmount := sdkmath.NewInt(99)

	// Create two epoch unbonding records, each with two host zone unbondings
	epochUnbondingRecords := []types.EpochUnbondingRecord{
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*types.HostZoneUnbonding{
				{HostZoneId: "chain-0", NativeTokenAmount: initialAmount},
				{HostZoneId: "chain-1", NativeTokenAmount: initialAmount},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*types.HostZoneUnbonding{
				{HostZoneId: "chain-0", NativeTokenAmount: initialAmount},
				{HostZoneId: "chain-1", NativeTokenAmount: initialAmount},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	// Update the amount for (epoch-1, chain-0) and (epoch-2, chain-1)
	updatedHostZoneUnbonding1 := types.HostZoneUnbonding{HostZoneId: "chain-0", NativeTokenAmount: updatedAmount}
	err := s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, 1, "chain-0", updatedHostZoneUnbonding1)
	s.Require().NoError(err, "no error expected when setting amount for (epoch-1, chain-0)")

	updatedHostZoneUnbonding2 := types.HostZoneUnbonding{HostZoneId: "chain-1", NativeTokenAmount: updatedAmount}
	err = s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, 2, "chain-1", updatedHostZoneUnbonding2)
	s.Require().NoError(err, "no error expected when setting amount for (epoch-2, chain-1)")

	// Create the mapping of expected native amounts
	expectedAmountMapping := map[uint64]map[string]sdkmath.Int{
		1: {
			"chain-0": updatedAmount,
			"chain-1": initialAmount,
		},
		2: {
			"chain-0": initialAmount,
			"chain-1": updatedAmount,
		},
	}

	// Loop the records and check that the amounts match the updates
	for _, epochUnbondingRecord := range s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx) {
		s.Require().Len(epochUnbondingRecord.HostZoneUnbondings, 2, "there should be two host records per epoch record")

		for _, hostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			expectedAmount := expectedAmountMapping[epochUnbondingRecord.EpochNumber][hostZoneUnbondingRecord.HostZoneId]
			s.Require().Equal(expectedAmount.Int64(), hostZoneUnbondingRecord.NativeTokenAmount.Int64(), "updated record amount")
		}
	}

}
