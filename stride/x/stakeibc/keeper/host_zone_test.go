package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/testutil/nullify"
	"github.com/Stride-labs/stride/x/stakeibc/keeper"
	"github.com/Stride-labs/stride/x/stakeibc/types"
)

func createTestHostZone(keeper *keeper.Keeper, ctx sdk.Context) types.HostZone {
	item := types.HostZone{}
	keeper.SetHostZone(ctx, item)
	return item
}

func TestHostZoneGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	item := createTestHostZone(keeper, ctx)
	rst, found := keeper.GetHostZone(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestHostZoneRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestHostZone(keeper, ctx)
	keeper.RemoveHostZone(ctx)
	_, found := keeper.GetHostZone(ctx)
	require.False(t, found)
}
