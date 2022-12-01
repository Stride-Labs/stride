package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Stride-Labs/stride/v3/testutil/keeper"
	"github.com/Stride-Labs/stride/v3/x/app-router/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.AppRouterKeeper(t)
	params := types.DefaultParams()
	params.Active = true

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
