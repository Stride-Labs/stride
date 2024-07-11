package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

    keepertest "github.com/Stride-Labs/stride/v22/testutil/keeper"
    "github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.AirdropKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
