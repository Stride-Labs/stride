package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Returns a new metric at the current block time and height
func NewMetric(ctx sdk.Context, key, value, metricType, attributes string) Metric {
	return Metric{
		Key:         key,
		Value:       value,
		MetricType:  metricType,
		UpdateTime:  ctx.BlockTime().Unix(),
		BlockHeight: ctx.BlockHeight(),
		Attributes:  attributes,
	}
}

// Returns the ID for a metric
func (m Metric) GetMetricID() string {
	return fmt.Sprintf("%s-%s-%d-%s", m.Key, m.Value, m.UpdateTime, m.DestinationOracle)
}
