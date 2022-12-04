package stakeibc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/stakeibc"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		IcaAccount: &types.ICAAccount{
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
	require.Equal(t, genesisState.IcaAccount, got.IcaAccount)
	require.Equal(t, genesisState.EpochTrackerList, got.EpochTrackerList)
	require.Equal(t, genesisState.Params, got.Params)
	// this line is used by starport scaffolding # genesis/test/assert
}
