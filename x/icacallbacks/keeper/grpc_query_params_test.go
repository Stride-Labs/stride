package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Stride-Labs/stride/v26/testutil/keeper"
	"github.com/Stride-Labs/stride/v26/x/icacallbacks/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.IcacallbacksKeeper(t)

	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
