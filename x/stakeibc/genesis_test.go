package stakeibc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

<<<<<<< HEAD
	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	"github.com/Stride-Labs/stride/v5/x/stakeibc"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
=======
	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	"github.com/Stride-Labs/stride/v5/x/stakeibc"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
>>>>>>> main
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		EpochTrackerList: []types.EpochTracker{
			{
				EpochIdentifier: "stride_epoch",
			},
		},
		HostZoneCount: sdk.ZeroInt(),
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.StakeibcKeeper(t)
	stakeibc.InitGenesis(ctx, *k, genesisState)
	got := stakeibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)
	require.Equal(t, genesisState.EpochTrackerList, got.EpochTrackerList)
	require.Equal(t, genesisState.Params, got.Params)
	// this line is used by starport scaffolding # genesis/test/assert
}
