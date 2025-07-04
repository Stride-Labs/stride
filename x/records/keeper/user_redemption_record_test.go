package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	keepertest "github.com/Stride-Labs/stride/v27/testutil/keeper"
	"github.com/Stride-Labs/stride/v27/testutil/nullify"
	"github.com/Stride-Labs/stride/v27/x/records/keeper"
	"github.com/Stride-Labs/stride/v27/x/records/types"
)

// TODO [cleanup]: Migrate to new KeeperTestSuite framework and remove use of nullify
func createNUserRedemptionRecord(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.UserRedemptionRecord {
	items := make([]types.UserRedemptionRecord, n)
	for i := range items {
		items[i].Id = strconv.Itoa(i)
		items[i].NativeTokenAmount = sdkmath.NewInt(int64(i))
		items[i].StTokenAmount = sdkmath.NewInt(int64(i))
		keeper.SetUserRedemptionRecord(ctx, items[i])
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
	actual := keeper.GetAllUserRedemptionRecord(ctx)
	require.Equal(t, len(items), len(actual))
}
