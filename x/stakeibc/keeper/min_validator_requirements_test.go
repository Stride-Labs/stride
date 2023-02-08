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

func createTestMinValidatorRequirements(keeper *keeper.Keeper, ctx sdk.Context) types.MinValidatorRequirements {
	item := types.MinValidatorRequirements{}
	keeper.SetMinValidatorRequirements(ctx, item)
	return item
}

func TestMinValidatorRequirementsGet(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	item := createTestMinValidatorRequirements(keeper, ctx)
	rst, found := keeper.GetMinValidatorRequirements(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestMinValidatorRequirementsRemove(t *testing.T) {
	keeper, ctx := keepertest.StakeibcKeeper(t)
	createTestMinValidatorRequirements(keeper, ctx)
	keeper.RemoveMinValidatorRequirements(ctx)
	_, found := keeper.GetMinValidatorRequirements(ctx)
	require.False(t, found)
}
