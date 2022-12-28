package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func createTestDelegation(keeper *keeper.Keeper, ctx sdk.Context) types.Delegation {
	item := types.Delegation{Amt: sdk.NewInt(1)}
	keeper.SetDelegation(ctx, item)
	return item
}

func TestDelegationGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	expected := createTestDelegation(keeper, ctx)
	actual, found := keeper.GetDelegation(ctx)
	fmt.Println("actual", actual)
	fmt.Println("expected", expected)
	require.True(t, found)

	// We use (gogoproto.nullable) for Int so when encode an empty delegation,
	// the marshaller will set Amount to zero, so the exist test below will false

	require.Equal(t,
		nullify.Fill(expected),
		nullify.Fill(actual),
	)
}

func TestDelegationRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestDelegation(keeper, ctx)
	keeper.RemoveDelegation(ctx)
	_, found := keeper.GetDelegation(ctx)
	require.False(t, found)
}
