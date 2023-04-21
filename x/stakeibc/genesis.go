package stakeibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, hostZone := range genState.HostZoneList {
		k.SetHostZone(ctx, hostZone)
	}
	for _, epochTracker := range genState.EpochTrackerList {
		k.SetEpochTracker(ctx, epochTracker)
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.HostZoneList = k.GetAllHostZone(ctx)
	genesis.EpochTrackerList = k.GetAllEpochTracker(ctx)

	return genesis
}
