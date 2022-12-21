package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func createTestValidator(keeper *keeper.Keeper, ctx sdk.Context) types.Validator {
	item := types.Validator{DelegationAmt: sdk.NewInt(1)}
	keeper.SetValidator(ctx, item)
	return item
}

func TestValidatorGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	item := createTestValidator(keeper, ctx)
	rst, found := keeper.GetValidator(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestValidatorRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestValidator(keeper, ctx)
	keeper.RemoveValidator(ctx)
	_, found := keeper.GetValidator(ctx)
	require.False(t, found)
}
