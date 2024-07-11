package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

    keepertest "github.com/Stride-Labs/stride/v22/testutil/keeper"
    "github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := keepertest.AirdropKeeper(t)
	params := types.DefaultParams()
	require.NoError(t, keeper.SetParams(ctx, params))

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
