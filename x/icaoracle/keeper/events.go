package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// Emits an event for an oracle update
func EmitUpdateOracleEvent(ctx sdk.Context, metric types.Metric) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateOracle,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyOracleChainId, metric.DestinationOracle),
			sdk.NewAttribute(types.AttributeKeyMetricID, metric.GetMetricID()),
			sdk.NewAttribute(types.AttributeKeyMetricKey, metric.Key),
			sdk.NewAttribute(types.AttributeKeyMetricValue, metric.Value),
			sdk.NewAttribute(types.AttributeKeyMetricType, metric.MetricType),
			sdk.NewAttribute(types.AttributeKeyMetricUpdateTime, fmt.Sprintf("%d", metric.UpdateTime)),
			sdk.NewAttribute(types.AttributeKeyMetricBlockHeight, fmt.Sprintf("%d", metric.BlockHeight)),
		),
	)
}

// Emits an event for an oracle update
func EmitUpdateOracleAckEvent(ctx sdk.Context, metric *types.Metric, ackStatus string) {
	if metric == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateOracleAck,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyOracleChainId, metric.DestinationOracle),
			sdk.NewAttribute(types.AttributeKeyMetricID, metric.GetMetricID()),
			sdk.NewAttribute(types.AttributeKeyMetricAckStatus, ackStatus),
			sdk.NewAttribute(types.AttributeKeyMetricKey, metric.Key),
			sdk.NewAttribute(types.AttributeKeyMetricValue, metric.Value),
			sdk.NewAttribute(types.AttributeKeyMetricType, metric.MetricType),
			sdk.NewAttribute(types.AttributeKeyMetricUpdateTime, fmt.Sprintf("%d", metric.UpdateTime)),
			sdk.NewAttribute(types.AttributeKeyMetricBlockHeight, fmt.Sprintf("%d", metric.BlockHeight)),
		),
	)
}
