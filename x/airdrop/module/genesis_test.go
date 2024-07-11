package airdrop_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/v22/testutil/keeper"
	"github.com/Stride-Labs/stride/v22/testutil/nullify"
	airdrop "github.com/Stride-Labs/stride/v22/x/airdrop/module"
	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AirdropKeeper(t)
	airdrop.InitGenesis(ctx, k, genesisState)
	got := airdrop.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	

	// this line is used by starport scaffolding # genesis/test/assert
}
