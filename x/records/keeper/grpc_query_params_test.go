package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.RecordsKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
