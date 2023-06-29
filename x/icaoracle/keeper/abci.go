package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// EndBlocker of icaoracle module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// For each queued metric, submit an ICA to each oracle, and then flag the metric as IN_PROGRESS
	for _, metric := range k.GetAllQueuedMetrics(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Submitting oracle metric update - Metric: %s, Oracle: %s", metric.Key, metric.DestinationOracle))

		// Ignore any inactive oracles
		oracle, found := k.GetOracle(ctx, metric.DestinationOracle)
		if !found || !oracle.Active {
			k.Logger(ctx).Info(fmt.Sprintf("Oracle %s is inactive", oracle.ChainId))
			continue
		}

		// Flag the metric as IN_PROGRESS to prevent resubmissions next block
		// We do this even in the case where the ICA submission fails (from something like a channel closure)
		// If the channel closes, once it is restored, the metric will get re-queued
		k.UpdateMetricStatus(ctx, metric, types.MetricStatus_IN_PROGRESS)

		if !k.IsOracleICAChannelOpen(ctx, oracle) {
			k.Logger(ctx).Error(fmt.Sprintf("Oracle %s has a closed ICA channel (%s)", oracle.ChainId, oracle.ChannelId))
			continue
		}

		// Submit the ICA for each metric
		err := k.SubmitMetricUpdate(ctx, oracle, metric)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to submit a metric update ICA - Metric: %+v, Oracle: %+v, %s", metric, oracle, err.Error()))
			continue
		}

		k.Logger(ctx).Info(fmt.Sprintf("Submitted metric update ICA - Metric: %s, Oracle: %s, Time: %d", metric.Key, oracle.ChainId, metric.UpdateTime))
	}
}
