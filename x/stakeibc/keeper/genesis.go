package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	for _, hostZone := range genState.HostZoneList {
		k.SetHostZone(ctx, hostZone)
	}
	for _, epochTracker := range genState.EpochTrackerList {
		k.SetEpochTracker(ctx, epochTracker)
	}
	for _, tradeRoute := range genState.TradeRoutes {
		k.SetTradeRoute(ctx, tradeRoute)
	}

	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)
	genesis.HostZoneList = k.GetAllHostZone(ctx)
	genesis.EpochTrackerList = k.GetAllEpochTracker(ctx)
	genesis.TradeRoutes = k.GetAllTradeRoutes(ctx)

	return genesis
}
