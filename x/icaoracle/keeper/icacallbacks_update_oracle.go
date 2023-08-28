package keeper

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v14/utils"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// Callback after an update oracle ICA
//
//	If successful/failure: the metric is removed from the pending store
//	If timeout: metric is left in pending store so it can be re-submitted
func (k Keeper) UpdateOracleCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	updateOracleCallback := types.UpdateOracleCallback{}
	if err := proto.Unmarshal(args, &updateOracleCallback); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal update oracle callback")
	}
	chainId := updateOracleCallback.OracleChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_UpdateOracle, "Starting update oracle callback"))

	// If the ack timed-out, log the error and exit successfully
	// The metric should remain in the pending store so that the ICA can be resubmitted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		EmitUpdateOracleAckEvent(ctx, updateOracleCallback.Metric, "timeout")
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_UpdateOracle, ackResponse.Status, packet))
		return nil
	}

	// if the ack fails, log the response as an error, otherwise log the success as an info log
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		EmitUpdateOracleAckEvent(ctx, updateOracleCallback.Metric, "failure")
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_UpdateOracle, ackResponse.Status, packet))
	} else {
		EmitUpdateOracleAckEvent(ctx, updateOracleCallback.Metric, "success")
		k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_UpdateOracle, ackResponse.Status, packet))
	}

	// Confirm the callback has a valid metric
	if updateOracleCallback.Metric == nil || updateOracleCallback.Metric.Key == "" {
		return errorsmod.Wrapf(types.ErrInvalidCallback, "metric is missing from callback: %+v", updateOracleCallback)
	}

	// Remove the metric from the store (aka mark update as complete)
	k.RemoveMetric(ctx, updateOracleCallback.Metric.GetMetricID())

	return nil
}
