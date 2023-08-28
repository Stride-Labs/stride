package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := genState.Validate(); err != nil {
		panic(err)
	}
	for _, oracle := range genState.Oracles {
		k.SetOracle(ctx, oracle)
	}
	for _, metric := range genState.Metrics {
		k.SetMetric(ctx, metric)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Oracles = k.GetAllOracles(ctx)
	genesis.Metrics = k.GetAllMetrics(ctx)

	return genesis
}
