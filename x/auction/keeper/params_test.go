package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.AuctionKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	p, _ := k.GetParams(ctx)
	require.EqualValues(t, params, p)
}
