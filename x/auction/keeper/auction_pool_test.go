package keeper_test

import (
	"testing"

    "github.com/Stride-Labs/stride/v5/x/auction/keeper"
    "github.com/Stride-Labs/stride/v5/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	"github.com/stretchr/testify/require"
)

func createNAuctionPool(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.AuctionPool {
	items := make([]types.AuctionPool, n)
	for i := range items {
		items[i].Id = keeper.AppendAuctionPool(ctx, items[i])
	}
	return items
}

func TestAuctionPoolGet(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	items := createNAuctionPool(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetAuctionPool(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestAuctionPoolRemove(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	items := createNAuctionPool(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveAuctionPool(ctx, item.Id)
		_, found := keeper.GetAuctionPool(ctx, item.Id)
		require.False(t, found)
	}
}

func TestAuctionPoolGetAll(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	items := createNAuctionPool(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllAuctionPool(ctx)),
	)
}

func TestAuctionPoolCount(t *testing.T) {
	keeper, ctx := keepertest.AuctionKeeper(t)
	items := createNAuctionPool(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetAuctionPoolCount(ctx))
}
