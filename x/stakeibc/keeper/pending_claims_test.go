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

func createNPendingClaims(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.PendingClaims {
	items := make([]types.PendingClaims, n)
	for i := range items {
		items[i].Sequence = strconv.Itoa(i)
        
		keeper.SetPendingClaims(ctx, items[i])
	}
	return items
}

func TestPendingClaimsGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNPendingClaims(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetPendingClaims(ctx,
		    item.Sequence,
            
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestPendingClaimsRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNPendingClaims(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemovePendingClaims(ctx,
		    item.Sequence,
            
		)
		_, found := keeper.GetPendingClaims(ctx,
		    item.Sequence,
            
		)
		require.False(t, found)
	}
}

func TestPendingClaimsGetAll(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	items := createNPendingClaims(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllPendingClaims(ctx)),
	)
}
