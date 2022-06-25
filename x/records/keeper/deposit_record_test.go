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

func createNDepositRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.DepositRecord {
	items := make([]types.DepositRecord, n)
	for i := range items {
		items[i].Id = keeper.AppendDepositRecord(ctx, items[i])
	}
	return items
}

func TestDepositRecordGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNDepositRecord(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetDepositRecord(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestDepositRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNDepositRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveDepositRecord(ctx, item.Id)
		_, found := keeper.GetDepositRecord(ctx, item.Id)
		require.False(t, found)
	}
}

func TestDepositRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNDepositRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllDepositRecord(ctx)),
	)
}

func TestDepositRecordCount(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNDepositRecord(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetDepositRecordCount(ctx))
}
