package keeper_test

import (
	"strconv"
	"testing"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNControllerBalances(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.ControllerBalances {
	items := make([]types.ControllerBalances, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)
        
		keeper.SetControllerBalances(ctx, items[i])
	}
	return items
}

func TestControllerBalancesGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNControllerBalances(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetControllerBalances(ctx,
		    item.Index,
            
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestControllerBalancesRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNControllerBalances(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveControllerBalances(ctx,
		    item.Index,
            
		)
		_, found := keeper.GetControllerBalances(ctx,
		    item.Index,
            
		)
		require.False(t, found)
	}
}

func TestControllerBalancesGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNControllerBalances(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllControllerBalances(ctx)),
	)
}
