package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v3/testutil/keeper"
	"github.com/Stride-Labs/stride/v3/testutil/nullify"
	"github.com/Stride-Labs/stride/v3/x/records/keeper"
	"github.com/Stride-Labs/stride/v3/x/records/types"
)

func createNEpochUnbondingRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.EpochUnbondingRecord {
	items := make([]types.EpochUnbondingRecord, n)
	for i, item := range items {
		item.EpochNumber = uint64(i)
		items[i] = item
		keeper.SetEpochUnbondingRecord(ctx, item)
	}
	return items
}

func TestEpochUnbondingRecordGet(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
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
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveEpochUnbondingRecord(ctx, item.EpochNumber)
		_, found := keeper.GetEpochUnbondingRecord(ctx, item.EpochNumber)
		require.False(t, found)
	}
}

func TestEpochUnbondingRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllEpochUnbondingRecord(ctx)),
	)
}

func TestGetAllPreviousEpochUnbondingRecords(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
	currentEpoch := uint64(8)
	fetchedItems := items[:currentEpoch]
	require.ElementsMatch(t,
		nullify.Fill(fetchedItems),
		nullify.Fill(keeper.GetAllPreviousEpochUnbondingRecords(ctx, currentEpoch)),
	)
}
