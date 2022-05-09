package keeper_test

import (
	"testing"

	testkeeper "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/x/epochs/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.EpochsKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
