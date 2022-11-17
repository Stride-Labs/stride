package epochs_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/v2/testutil/keeper"
	"github.com/Stride-Labs/stride/v2/testutil/nullify"
	"github.com/Stride-Labs/stride/v2/x/epochs"
	"github.com/Stride-Labs/stride/v2/x/epochs/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
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
