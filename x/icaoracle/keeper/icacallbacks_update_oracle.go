package keeper

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v5/utils"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Callback after an update oracle ICA
// Removes metric from pending store (regardless of ack status)
func UpdateOracleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	updateOracleCallback := types.UpdateOracleCallback{}
	if err := proto.Unmarshal(args, &updateOracleCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal update oracle callback: %s", err.Error())
	}
	chainId := updateOracleCallback.OracleChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_UpdateOracle, "Starting update oracle callback"))

	// Log ack status
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT ||
		ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_UpdateOracle, ackResponse.Status, packet))
	}
	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_UpdateOracle,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Confirm the callback has a valid metric
	if updateOracleCallback.Metric == nil || updateOracleCallback.Metric.Key == "" {
		return errorsmod.Wrapf(types.ErrInvalidCallback, "metric is missing from callback: %+v", updateOracleCallback)
	}

	// Remove the metric from the pending store (aka mark update as complete)
	k.SetMetricUpdateComplete(ctx, updateOracleCallback.Metric.Key, updateOracleCallback.OracleChainId)

	return nil
}
