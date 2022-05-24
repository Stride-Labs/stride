package keeper_test

import (
	"testing"

	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func createTestICAAccount(keeper *keeper.Keeper, ctx sdk.Context) types.ICAAccount {
	item := types.ICAAccount{}
	keeper.SetICAAccount(ctx, item)
	return item
}

func TestICAAccountGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	item := createTestICAAccount(keeper, ctx)
	rst, found := keeper.GetICAAccount(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestICAAccountRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestICAAccount(keeper, ctx)
	keeper.RemoveICAAccount(ctx)
	_, found := keeper.GetICAAccount(ctx)
	require.False(t, found)
}
