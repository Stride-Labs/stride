package liquidgov_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	"github.com/Stride-Labs/stride/v5/x/liquidgov"
	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.LiquidgovKeeper(t)
	liquidgov.InitGenesis(ctx, *k, genesisState)
	got := liquidgov.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	

	// this line is used by starport scaffolding # genesis/test/assert
}
