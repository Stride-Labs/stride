package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// Stores a metric in the main metric store and then either
// adds the metric to the queue or removes it from the queue
// depending on the status of the metric
func (k Keeper) SetMetric(ctx sdk.Context, metric types.Metric) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricKeyPrefix)

	metricKey := types.KeyPrefix(metric.GetMetricID())
	metricValue := k.cdc.MustMarshal(&metric)

	store.Set(metricKey, metricValue)

	switch metric.Status {
	case types.MetricStatus_QUEUED:
		k.addMetricToQueue(ctx, metricKey)
	case types.MetricStatus_IN_PROGRESS:
		k.removeMetricFromQueue(ctx, metricKey)
	default:
		panic("metric status must be specified as QUEUED or IN_PROGRESS before storing")
	}
}

// Gets a specifc metric from the store
func (k Keeper) GetMetric(ctx sdk.Context, metricId string) (metric types.Metric, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricKeyPrefix)

	metricKey := types.KeyPrefix(metricId)
	metricBz := store.Get(metricKey)

	if len(metricBz) == 0 {
		return metric, false
	}

	k.cdc.MustUnmarshal(metricBz, &metric)
	return metric, true
}

// Returns all metrics from the store
func (k Keeper) GetAllMetrics(ctx sdk.Context) (metrics []types.Metric) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {

		metric := types.Metric{}
		k.cdc.MustUnmarshal(iterator.Value(), &metric)
		metrics = append(metrics, metric)
	}

	return metrics
}

// Removes a metric from the store
func (k Keeper) RemoveMetric(ctx sdk.Context, metricId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricKeyPrefix)
	metricKey := types.KeyPrefix(metricId)
	store.Delete(metricKey)
	k.removeMetricFromQueue(ctx, metricKey)
}

// Updates the status of a metric which will consequently move it either
// in or out of the queue
func (k Keeper) UpdateMetricStatus(ctx sdk.Context, metric types.Metric, status types.MetricStatus) {
	metric.Status = status
	k.SetMetric(ctx, metric)
}

// Adds a metric to the queue, which acts as an index for all metrics
// that should be submitted to it's relevant oracle
func (k Keeper) addMetricToQueue(ctx sdk.Context, metricKey []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)
	store.Set(metricKey, []byte{1})
}

// Removes a metric from the queue
func (k Keeper) removeMetricFromQueue(ctx sdk.Context, metricKey []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)
	store.Delete(metricKey)
}

// Returns all metrics from the index queue
func (k Keeper) GetAllQueuedMetrics(ctx sdk.Context) (metrics []types.Metric) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricQueueKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {

		metricId := string(iterator.Key())
		metric, found := k.GetMetric(ctx, metricId)
		if !found {
			panic("metric in queue but not metric store")
		}
		metrics = append(metrics, metric)
	}

	return metrics
}
