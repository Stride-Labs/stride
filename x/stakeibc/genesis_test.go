package stakeibc_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/v2/testutil/keeper"
	"github.com/Stride-Labs/stride/v2/testutil/nullify"
	"github.com/Stride-Labs/stride/v2/x/stakeibc"
	"github.com/Stride-Labs/stride/v2/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		ICAAccount: &types.ICAAccount{
			Address: "78",
		},
		EpochTrackerList: []types.EpochTracker{
			{EpochIdentifier: "stride_epoch"},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.StakeibcKeeper(t)
	stakeibc.InitGenesis(ctx, *k, genesisState)
	got := stakeibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)
	require.Equal(t, genesisState.ICAAccount, got.ICAAccount)
	require.Equal(t, genesisState.EpochTrackerList, got.EpochTrackerList)
	require.Equal(t, genesisState.Params, got.Params)
	// this line is used by starport scaffolding # genesis/test/assert
}
