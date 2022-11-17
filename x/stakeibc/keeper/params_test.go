package keeper_test

import (
	"testing"

	testkeeper "github.com/Stride-Labs/stride/v2/testutil/keeper"
	"github.com/Stride-Labs/stride/v2/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.StakeibcKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
