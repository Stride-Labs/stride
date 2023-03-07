package icaoracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/keeper"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, oracle := range genState.Oracles {
		k.SetOracle(ctx, oracle)
	}
	for _, queuedMetric := range genState.QueuedMetrics {
		k.QueueMetricUpdate(ctx, queuedMetric)
	}
	for _, pendingMetricUpdate := range genState.PendingMetrics {
		k.SetMetricUpdateInProgress(ctx, pendingMetricUpdate)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	oracles := k.GetAllOracles(ctx)
	queuedMetrics := k.GetAllMetricsFromQueue(ctx)
	pendingMetrics := k.GetAllPendingMetricUpdates(ctx)

	genesis.Oracles = oracles
	genesis.QueuedMetrics = queuedMetrics
	genesis.PendingMetrics = pendingMetrics

	return genesis
}
