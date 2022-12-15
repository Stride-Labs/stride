package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/records/keeper"
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func createNEpochUnbondingRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) ([]types.EpochUnbondingRecord, map[string]types.HostZoneUnbonding) {
	hostZoneUnbondingsList := []types.HostZoneUnbonding{
		{
			HostZoneId: "host-A",
			Status:     types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount: sdk.ZeroInt(),
			NativeTokenAmount: sdk.ZeroInt(),
		},
		{
			HostZoneId: "host-B",
			Status:     types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount: sdk.ZeroInt(),
			NativeTokenAmount: sdk.ZeroInt(),
		},
		{
			HostZoneId: "host-C",
			Status:     types.HostZoneUnbonding_UNBONDING_QUEUE,
			StTokenAmount: sdk.ZeroInt(),
			NativeTokenAmount: sdk.ZeroInt(),
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
