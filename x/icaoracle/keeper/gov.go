package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"
)

// Proposal handler for toggling whether an oracle is currently active (meaning it's a destination for metric pushes)
func (k Keeper) HandleToggleOracleProposal(ctx sdk.Context, proposal *types.ToggleOracleProposal) error {
	return k.ToggleOracle(ctx, proposal.OracleChainId, proposal.Active)
}

// Proposal handler for removing an oracle from the store
func (k Keeper) HandleRemoveOracleProposal(ctx sdk.Context, proposal *types.RemoveOracleProposal) error {
	_, found := k.GetOracle(ctx, proposal.OracleChainId)
	if !found {
		return types.ErrOracleNotFound
	}

	k.RemoveOracle(ctx, proposal.OracleChainId)

	// Remove all metrics that were targeting this oracle
	for _, metric := range k.GetAllMetrics(ctx) {
		if metric.DestinationOracle == proposal.OracleChainId {
			k.RemoveMetric(ctx, metric.GetMetricID())
		}
	}

	return nil
}
