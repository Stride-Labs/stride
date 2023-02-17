package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// --------------------------------------------------
//       METRICS QUEUED FOR ICA SUBMISSION
// --------------------------------------------------

// Adds a metric update to the queue (i.e. adds to the queue store)
// The metrics are stored using the metric key, so if the same metric is already
// in the store, it will get overridden
func (k Keeper) QueueMetricUpdate(ctx sdk.Context, metric types.Metric) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)

	metricKey := types.KeyPrefix(metric.Key)
	metricValue := k.cdc.MustMarshal(&metric)

	store.Set(metricKey, metricValue)
}

// Gets a specifc metric from the queue
func (k Keeper) GetMetricFromQueue(ctx sdk.Context, key string) (metric types.Metric, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)

	metricKey := types.KeyPrefix(key)
	metricBz := store.Get(metricKey)

	if len(metricBz) == 0 {
		return metric, false
	}

	k.cdc.MustUnmarshal(metricBz, &metric)
	return metric, true
}

// Gets all metrics from the queue
func (k Keeper) GetAllMetricsFromQueue(ctx sdk.Context) []types.Metric {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allQueuedMetrics := []types.Metric{}
	for ; iterator.Valid(); iterator.Next() {

		metric := types.Metric{}
		k.cdc.MustUnmarshal(iterator.Value(), &metric)
		allQueuedMetrics = append(allQueuedMetrics, metric)
	}

	return allQueuedMetrics
}

// Removes a metric from the queue (i.e. removes from the queue store)
func (k Keeper) RemoveMetricFromQueue(ctx sdk.Context, key string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)
	metricKey := types.KeyPrefix(key)
	store.Delete(metricKey)
}

// --------------------------------------------------
//          METRICS WITH ICA's IN PROGRESS
// --------------------------------------------------

// Marks a metric update as "in progress" which moves it to the pending store
// The pending store has one record per ICA call, meaning the same metric
// will be stored multiple times in the pending store if there are multiple oracles
func (k Keeper) SetMetricUpdateInProgress(ctx sdk.Context, pendingMetric types.PendingMetricUpdate) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)

	pendingMetricKey := types.GetPendingMetricKey(pendingMetric.Metric.Key, pendingMetric.OracleMoniker)
	pendingMetricBz := k.cdc.MustMarshal(&pendingMetric)

	store.Set(pendingMetricKey, pendingMetricBz)
}

// Gets a pending metric update from the pending store
func (k Keeper) GetPendingMetricUpdate(ctx sdk.Context, key string, oracleMoniker string) (pendingMetric types.PendingMetricUpdate, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)

	pendingMetricKey := types.GetPendingMetricKey(key, oracleMoniker)
	pendingMetricBz := store.Get(pendingMetricKey)

	if len(pendingMetricBz) == 0 {
		return pendingMetric, false
	}

	k.cdc.MustUnmarshal(pendingMetricBz, &pendingMetric)
	return pendingMetric, true
}

// Gets all pending metric updates from the pending store
func (k Keeper) GetAllPendingMetricUpdates(ctx sdk.Context) []types.PendingMetricUpdate {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allPendingMetricUpdates := []types.PendingMetricUpdate{}
	for ; iterator.Valid(); iterator.Next() {

		pendingMetric := types.PendingMetricUpdate{}
		k.cdc.MustUnmarshal(iterator.Value(), &pendingMetric)
		allPendingMetricUpdates = append(allPendingMetricUpdates, pendingMetric)
	}

	return allPendingMetricUpdates
}

// Marks a metric update as "complete", meaning it has been updated on the oracle
// and the ack has been received. Indicated by removing it from the pending store
func (k Keeper) SetMetricUpdateComplete(ctx sdk.Context, metricKey string, oracleMoniker string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)
	metricUpdateKey := types.GetPendingMetricKey(metricKey, oracleMoniker)
	store.Delete(metricUpdateKey)
}
