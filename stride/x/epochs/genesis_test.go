package epochs_test

import (
	"testing"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/testutil/nullify"
	"github.com/Stride-labs/stride/x/epochs"
	"github.com/Stride-labs/stride/x/epochs/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.EpochsKeeper(t)
	epochs.InitGenesis(ctx, *k, genesisState)
	got := epochs.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
