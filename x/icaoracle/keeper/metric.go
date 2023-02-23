package keeper

import (
	"encoding/json"
	"time"

	errorsmod "cosmossdk.io/errors"
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

// Submits an ICA to update the metric in the CW contract
func (k Keeper) SubmitMetricUpdate(ctx sdk.Context, oracle types.Oracle, metric types.Metric) error {
	// Validate ICA is setup properly, contract has been instantiated, and oracle is active
	if err := oracle.ValidateICASetup(); err != nil {
		return err
	}
	if err := oracle.ValidateContractInstantiated(); err != nil {
		return err
	}
	if !oracle.Active {
		return errorsmod.Wrapf(types.ErrOracleInactive, "oracle (%s) is inactive", oracle.ChainId)
	}

	// Build contract message with metric update
	contractMsg := types.MsgExecuteContractUpdateMetric{
		UpdateMetric: &metric,
	}
	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to marshal execute contract update metric: %s", err.Error())
	}

	// Build ICA message to execute the CW contract
	msgs := []sdk.Msg{&types.MsgExecuteContract{
		Sender:   oracle.IcaAddress,
		Contract: oracle.ContractAddress,
		Msg:      contractMsgBz,
	}}

	// QUESTION/TODO: Not sure what makes the most sense for the timeout
	// I think we can be more conservative than our epochly logic
	// The oracle querier can enforce filters to ensure the data is recent, so I think from the Stride
	// perspective, we should lean more conservative and do our best to avoid timeout's and channel closure's
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour * 24).Nanoseconds())

	// Submit the ICA to execute the contract
	callbackArgs := types.UpdateOracleCallback{
		OracleChainId: oracle.ChainId,
		Metric:        &metric,
	}
	icaTx := types.ICATx{
		ConnectionId: oracle.ConnectionId,
		ChannelId:    oracle.ChannelId,
		PortId:       oracle.PortId,
		Messages:     msgs,
		Timeout:      timeout,
		CallbackArgs: &callbackArgs,
		CallbackId:   ICACallbackID_UpdateOracle,
	}
	if err := k.SubmitICATx(ctx, icaTx); err != nil {
		return errorsmod.Wrapf(err, "unable to submit update oracle contract ICA")
	}

	// Add the metric to the pending store
	pendingMetricUpdate := types.PendingMetricUpdate{
		OracleChainId: oracle.ChainId,
		Metric:        &metric,
	}
	k.SetMetricUpdateInProgress(ctx, pendingMetricUpdate)

	return nil
}

// --------------------------------------------------
//          METRICS WITH ICA's IN PROGRESS
// --------------------------------------------------

// Marks a metric update as "in progress" which moves it to the pending store
// The pending store has one record per ICA call, meaning the same metric
// will be stored multiple times in the pending store if there are multiple oracles
func (k Keeper) SetMetricUpdateInProgress(ctx sdk.Context, pendingMetric types.PendingMetricUpdate) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)

	pendingMetricKey := types.GetPendingMetricKey(pendingMetric.Metric.Key, pendingMetric.OracleChainId)
	pendingMetricBz := k.cdc.MustMarshal(&pendingMetric)

	store.Set(pendingMetricKey, pendingMetricBz)
}

// Gets a pending metric update from the pending store
func (k Keeper) GetPendingMetricUpdate(ctx sdk.Context, key string, oracleChainId string) (pendingMetric types.PendingMetricUpdate, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)

	pendingMetricKey := types.GetPendingMetricKey(key, oracleChainId)
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
func (k Keeper) SetMetricUpdateComplete(ctx sdk.Context, metricKey string, oracleChainId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MetricPendingKeyPrefix)
	metricUpdateKey := types.GetPendingMetricKey(metricKey, oracleChainId)
	store.Delete(metricUpdateKey)
}
