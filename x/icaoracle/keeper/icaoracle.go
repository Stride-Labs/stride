package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// QUESTION: Not sure what makes the most sense for the timeout
// I think we can be more conservative than our epochly logic
// The oracle querier can enforce filters to ensure the data is recent, so I think from the Stride
// perspective, we should lean more conservative and do our best to avoid timeout's and channel closure's
var (
	MetricUpdateTimeout = time.Hour * 24 // 1 day
)

// Queues an metric update across each active oracle
// One metric record is created for each oracle, in status QUEUED
func (k Keeper) QueueMetricUpdate(ctx sdk.Context, key, value, metricType, attributes string) {
	metric := types.NewMetric(ctx, key, value, metricType, attributes)
	metric.Status = types.MetricStatus_METRIC_STATUS_QUEUED

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
	contractMsg := types.MsgExecuteContractPostMetric{
		PostMetric: &metric,
	}
	contractMsgBz, err := json.Marshal(contractMsg)
	if err != nil {
		return errorsmod.Wrapf(types.ErrMarshalFailure, "unable to marshal execute contract post metric: %s", err.Error())
	}

	// Build ICA message to execute the CW contract
	msgs := []sdk.Msg{&types.MsgExecuteContract{
		Sender:   oracle.IcaAddress,
		Contract: oracle.ContractAddress,
		Msg:      contractMsgBz,
	}}

	timeout := uint64(ctx.BlockTime().UnixNano() + MetricUpdateTimeout.Nanoseconds())

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

	return nil
}
