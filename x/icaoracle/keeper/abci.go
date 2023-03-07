package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// EndBlocker of icaoracle module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// QUESTION: Should we add metric count to the store so that we can do a simple lookup here
	// or will that be essentially just as expensive as calling GetAllMetricsFromQueue when the queue is empty?

	// For each queued metric, submit an ICA to each oracle, and then move the metric from
	// the queue to the pending store
	for _, latestMetric := range k.GetAllMetricsFromQueue(ctx) {
		for _, oracle := range k.GetAllOracles(ctx) {
			k.Logger(ctx).Info(fmt.Sprintf("Submitting oracle metric update - Metric: %s, Oracle: %s", latestMetric.Key, oracle.ChainId))

			// Ignore any inactive oracles
			if !oracle.Active {
				k.Logger(ctx).Info(fmt.Sprintf("Oracle %s is inactive", oracle.ChainId))
				continue
			}

			// Build the metric + oracle combo that will be used to track ICAs that are in flight or that need to be submitted later
			pendingMetricUpdate := types.PendingMetricUpdate{
				Metric:        &latestMetric,
				OracleChainId: oracle.ChainId,
			}

			// If the oracle is inactive, we'll still add the metric in the pending store so that
			// it can get submitted when the oracle comes back online
			if !k.IsOracleICAChannelOpen(ctx, oracle) {
				k.SetMetricUpdateInProgress(ctx, pendingMetricUpdate)
				k.Logger(ctx).Error(fmt.Sprintf("Oracle %s has a closed ICA channel (%s)", oracle.ChainId, oracle.ChannelId))
				continue
			}

			// In the event of a channel closure, some pending metrics can accrue
			// After the channel has been restored, the oracle will have a flag set that indicates when it's time to submit the older metrics
			olderMetrics := []types.Metric{}
			if oracle.FlushPendingMetrics {
				olderMetrics = k.GetPendingMetrics(ctx, latestMetric.Key, oracle.ChainId)

				// Set the flush flag back to false
				oracle.FlushPendingMetrics = false
				k.SetOracle(ctx, oracle)
			}

			// Submit the ICA for each metric
			for _, metric := range append(olderMetrics, latestMetric) {
				err := k.SubmitMetricUpdate(ctx, oracle, metric)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("Failed to submit a metric update ICA - Metric: %+v, Oracle: %+v, %s", metric, oracle, err.Error()))
				} else {
					k.Logger(ctx).Info(fmt.Sprintf("Submitted metric update ICA - Metric: %s, Oracle: %s, Time: %d", metric.Key, oracle.ChainId, metric.UpdateTime))
					k.SetMetricUpdateInProgress(ctx, pendingMetricUpdate)
				}
			}
		}
		k.RemoveMetricFromQueue(ctx, latestMetric.Key)
	}
}
