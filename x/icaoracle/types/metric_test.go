package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

// Tests NewMetric and GetMetricID
func TestMetric(t *testing.T) {
	blockHeight := int64(10)
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	blockTimeUnix := int64(1672531200)

	ctx := sdk.Context{}.
		WithBlockHeight(blockHeight).
		WithBlockTime(blockTime)

	key := "key"
	value := "value"
	attributes := "attributes"
	metricType := "type"

	expectedMetric := types.Metric{
		Key:         key,
		Value:       value,
		MetricType:  metricType,
		UpdateTime:  blockTimeUnix,
		BlockHeight: blockHeight,
		Attributes:  "attributes",
		Status:      types.MetricStatus_UNSPECIFIED,
	}

	actualMetric := types.NewMetric(ctx, key, value, metricType, attributes)
	require.Equal(t, expectedMetric, actualMetric, "metric")

	actualMetric.DestinationOracle = "chain"
	expectedId := "key-value-1672531200-chain"
	require.Equal(t, expectedId, actualMetric.GetMetricID(), "metric ID")
}
