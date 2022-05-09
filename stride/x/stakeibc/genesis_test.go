package stakeibc_test

import (
	"testing"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/testutil/nullify"
	"github.com/Stride-labs/stride/x/stakeibc"
	"github.com/Stride-labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		HostZone: &types.HostZone{
			PortId:    "76",
			ChannelId: "85",
		},
		ICAAccount: &types.ICAAccount{
			Address:          "78",
			Balance:          49,
			DelegatedBalance: 80,
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

	require.Equal(t, genesisState.HostZone, got.HostZone)
	require.Equal(t, genesisState.ICAAccount, got.ICAAccount)
	// this line is used by starport scaffolding # genesis/test/assert
}
