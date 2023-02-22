package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	keepertest "github.com/Stride-Labs/stride/v6/testutil/keeper"
	"github.com/Stride-Labs/stride/v6/testutil/nullify"
	"github.com/Stride-Labs/stride/v6/x/records/keeper"
	"github.com/Stride-Labs/stride/v6/x/records/types"
)

func createNEpochUnbondingRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) ([]types.EpochUnbondingRecord, map[string]types.HostZoneUnbonding) {
	hostZoneUnbondingsList := []types.HostZoneUnbonding{
		{
			HostZoneId:        "host-A",
			Status:            types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount:     sdkmath.ZeroInt(),
			NativeTokenAmount: sdkmath.ZeroInt(),
		},
		{
			HostZoneId:        "host-B",
			Status:            types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount:     sdkmath.ZeroInt(),
			NativeTokenAmount: sdkmath.ZeroInt(),
		},
		{
			HostZoneId:        "host-C",
			Status:            types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount:     sdkmath.ZeroInt(),
			NativeTokenAmount: sdkmath.ZeroInt(),
		},
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
		keeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	}
	return epochUnbondingRecords, hostZoneUnbondingsMap
}

func TestEpochUnbondingRecordGet(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetEpochUnbondingRecord(ctx, item.EpochNumber)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestEpochUnbondingRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveEpochUnbondingRecord(ctx, item.EpochNumber)
		_, found := keeper.GetEpochUnbondingRecord(ctx, item.EpochNumber)
		require.False(t, found)
	}
}

func TestEpochUnbondingRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllEpochUnbondingRecord(ctx)),
	)
}

func TestGetAllPreviousEpochUnbondingRecords(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items, _ := createNEpochUnbondingRecord(keeper, ctx, 10)
	currentEpoch := uint64(8)
	fetchedItems := items[:currentEpoch]
	require.ElementsMatch(t,
		nullify.Fill(fetchedItems),
		nullify.Fill(keeper.GetAllPreviousEpochUnbondingRecords(ctx, currentEpoch)),
	)
}

func TestGetHostZoneUnbondingByChainId(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	_, hostZoneUnbondings := createNEpochUnbondingRecord(keeper, ctx, 10)

	expectedHostZoneUnbonding := hostZoneUnbondings["host-B"]
	actualHostZoneUnbonding, found := keeper.GetHostZoneUnbondingByChainId(ctx, 1, "host-B")

	require.True(t, found)
	require.Equal(t,
		*actualHostZoneUnbonding,
		expectedHostZoneUnbonding,
	)
}

func TestAddHostZoneToEpochUnbondingRecord(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	epochUnbondingRecords, _ := createNEpochUnbondingRecord(keeper, ctx, 3)

	epochNumber := 0
	initialEpochUnbondingRecord := epochUnbondingRecords[epochNumber]

	// Add new host zone to initial epoch unbonding records
	newHostZone := types.HostZoneUnbonding{
		HostZoneId: "host-D",
		Status:     types.HostZoneUnbonding_UNBONDING_QUEUE,
	}
	expectedEpochUnbondingRecord := initialEpochUnbondingRecord
	expectedEpochUnbondingRecord.HostZoneUnbondings = append(expectedEpochUnbondingRecord.HostZoneUnbondings, &newHostZone)

	actualEpochUnbondingRecord, success := keeper.AddHostZoneToEpochUnbondingRecord(ctx, uint64(epochNumber), "host-D", &newHostZone)

	require.True(t, success)
	require.Equal(t,
		expectedEpochUnbondingRecord,
		*actualEpochUnbondingRecord,
	)
}

func TestSetHostZoneUnbondings(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)

	initialEpochUnbondingRecords, _ := createNEpochUnbondingRecord(keeper, ctx, 4)

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

	err := keeper.SetHostZoneUnbondings(ctx, hostIdToUpdate, epochsToUpdate, newStatus)
	require.Nil(t, err)

	actualEpochUnbondingRecord := keeper.GetAllEpochUnbondingRecord(ctx)
	require.ElementsMatch(t,
		expectedEpochUnbondingRecords,
		actualEpochUnbondingRecord,
	)
}

func (s *KeeperTestSuite) TestResetEpochUnbondingRecordEpochNumbers() {
	// Create 4 epoch unbonding records with epoch numbers 0,1,2, and 3
	initialEpochUnbondingRecords, _ := createNEpochUnbondingRecord(&s.App.RecordsKeeper, s.Ctx, 4)

	// Update each epoch unbonding record to have a 0 epoch number
	for _, epochUnbondingRecord := range initialEpochUnbondingRecords {
		recordsStore := s.Ctx.KVStore(s.App.GetKey(types.ModuleName))
		epochUnbondingRecordStore := prefix.NewStore(recordsStore, types.KeyPrefix(types.EpochUnbondingRecordKey))

		// Update the epoch number to 0 and reset the record with the old store key
		storedEpochNumberBz := keeper.GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.EpochNumber)
		epochUnbondingRecord.EpochNumber = 0

		epochUnbondingRecordBz, err := s.App.RecordsKeeper.Cdc.Marshal(&epochUnbondingRecord)
		s.Require().NoError(err, "there should be no error marshalling the epoch unbonding record")
		epochUnbondingRecordStore.Set(storedEpochNumberBz, epochUnbondingRecordBz)
	}

	// Confirm all epoch unbonding records are 0
	for i, epochUnbondingRecord := range s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx) {
		s.Require().Equal(uint64(0), epochUnbondingRecord.EpochNumber, "before the reset, epoch unbonding record %d should have EpochNumber 0", i)
	}

	// Reset epoch unbonding record numbers
	err := s.App.RecordsKeeper.ResetEpochUnbondingRecordEpochNumbers(s.Ctx)
	s.Require().NoError(err, "resetting the epoch unbonding record numbers should not error")

	// Config epoch unbonding records were updated
	for i, epochUnbondingRecord := range s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx) {
		s.Require().Equal(uint64(i), epochUnbondingRecord.EpochNumber, "after the reset, epoch unbonding record %d should have EpochNumber %d", i)
	}
}
