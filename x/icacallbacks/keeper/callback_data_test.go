package keeper_test

import (
	"strconv"
	"testing"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/icacallbacks/keeper"
	"github.com/Stride-Labs/stride/x/icacallbacks/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNCallbackData(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.CallbackData {
	items := make([]types.CallbackData, n)
	for i := range items {
		items[i].CallbackKey = strconv.Itoa(i)

		keeper.SetCallbackData(ctx, items[i])
	}
	return items
}

func TestCallbackDataGet(t *testing.T) {
	keeper, ctx := keepertest.IcacallbacksKeeper(t)
	items := createNCallbackData(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetCallbackData(ctx,
			item.CallbackKey,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestCallbackDataRemove(t *testing.T) {
	keeper, ctx := keepertest.IcacallbacksKeeper(t)
	items := createNCallbackData(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveCallbackData(ctx,
			item.CallbackKey,
		)
		_, found := keeper.GetCallbackData(ctx,
			item.CallbackKey,
		)
		require.False(t, found)
	}
}

func TestCallbackDataGetAll(t *testing.T) {
	keeper, ctx := keepertest.IcacallbacksKeeper(t)
	items := createNCallbackData(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllCallbackData(ctx)),
	)
}
