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

func createNUserRedemptionRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.UserRedemptionRecord {
	items := make([]types.UserRedemptionRecord, n)
	for i := range items {
		items[i].Id = keeper.AppendUserRedemptionRecord(ctx, items[i])
	}
	return items
}

func TestUserRedemptionRecordGet(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNUserRedemptionRecord(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetUserRedemptionRecord(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestUserRedemptionRecordRemove(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNUserRedemptionRecord(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveUserRedemptionRecord(ctx, item.Id)
		_, found := keeper.GetUserRedemptionRecord(ctx, item.Id)
		require.False(t, found)
	}
}

func TestUserRedemptionRecordGetAll(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNUserRedemptionRecord(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllUserRedemptionRecord(ctx)),
	)
}

func TestUserRedemptionRecordCount(t *testing.T) {
	keeper, ctx := keepertest.RecordsKeeper(t)
	items := createNUserRedemptionRecord(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetUserRedemptionRecordCount(ctx))
}
