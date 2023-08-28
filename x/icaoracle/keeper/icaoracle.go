package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

var (
	InstantiateOracleTimeout = time.Hour * 24 // 1 day
	MetricUpdateTimeout      = time.Hour * 24 // 1 day
)

// Queues an metric update across each active oracle
// One metric record is created for each oracle, in status QUEUED
func (k Keeper) QueueMetricUpdate(ctx sdk.Context, key, value, metricType, attributes string) {
	metric := types.NewMetric(ctx, key, value, metricType, attributes)
	metric.Status = types.MetricStatus_QUEUED

	for _, oracle := range k.GetAllOracles(ctx) {
		// Ignore any inactive oracles
		if !oracle.Active {
			continue
		}

		metric.DestinationOracle = oracle.ChainId
		k.SetMetric(ctx, metric)

		k.Logger(ctx).Info(fmt.Sprintf("Queueing oracle metric update - Metric: %s, Oracle: %s", metric.Key, oracle.ChainId))
	}
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
	contractMsg := types.NewMsgExecuteContractPostMetric(metric)
	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal execute contract post metric")
	}

	// Build ICA message to execute the CW contract
	msgs := []proto.Message{&types.MsgExecuteContract{
		Sender:   oracle.IcaAddress,
		Contract: oracle.ContractAddress,
		Msg:      contractMsgBz,
	}}

	// Submit the ICA to execute the contract
	callbackArgs := types.UpdateOracleCallback{
		OracleChainId: oracle.ChainId,
		Metric:        &metric,
	}
	icaTx := types.ICATx{
		ConnectionId:    oracle.ConnectionId,
		ChannelId:       oracle.ChannelId,
		PortId:          oracle.PortId,
		Owner:           types.FormatICAAccountOwner(oracle.ChainId, types.ICAAccountType_Oracle),
		Messages:        msgs,
		RelativeTimeout: MetricUpdateTimeout,
		CallbackArgs:    &callbackArgs,
		CallbackId:      ICACallbackID_UpdateOracle,
	}
	if err := k.SubmitICATx(ctx, icaTx); err != nil {
		return errorsmod.Wrapf(err, "unable to submit update oracle contract ICA")
	}

	return nil
}

// For each queued metric, submit an ICA to each oracle, and then flag the metric as IN_PROGRESS
func (k Keeper) PostAllQueuedMetrics(ctx sdk.Context) {
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
		if err := k.SubmitMetricUpdate(ctx, oracle, metric); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to submit a metric update ICA - Metric: %+v, Oracle: %+v, %s", metric, oracle, err.Error()))
			continue
		}

		k.Logger(ctx).Info(fmt.Sprintf("Submitted metric update ICA - Metric: %s, Oracle: %s, Time: %d", metric.Key, oracle.ChainId, metric.UpdateTime))
		EmitUpdateOracleEvent(ctx, metric)
	}
}
