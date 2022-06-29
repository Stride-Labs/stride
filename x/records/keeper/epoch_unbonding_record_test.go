package keeper_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/records/keeper"
	"github.com/Stride-Labs/stride/x/records/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func createNEpochUnbondingRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.EpochUnbondingRecord {
	items := make([]types.EpochUnbondingRecord, n)
	for i := range items {
		items[i].Id = keeper.AppendEpochUnbondingRecord(ctx, items[i])
	}
	return items
}

func TestEpochUnbondingRecordGet(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetEpochUnbondingRecord(ctx, item.Id)
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
		keeper.RemoveEpochUnbondingRecord(ctx, item.Id)
		_, found := keeper.GetEpochUnbondingRecord(ctx, item.Id)
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

func TestEpochUnbondingRecordCount(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNEpochUnbondingRecord(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetEpochUnbondingRecordCount(ctx))
}
