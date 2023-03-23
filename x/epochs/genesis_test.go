package epochs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v7/testutil/keeper"
	"github.com/Stride-Labs/stride/v7/testutil/nullify"
	"github.com/Stride-Labs/stride/v7/x/epochs"
	"github.com/Stride-Labs/stride/v7/x/epochs/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{}

	k, ctx := keepertest.EpochsKeeper(t)
	epochs.InitGenesis(ctx, *k, genesisState)
	got := epochs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

}
