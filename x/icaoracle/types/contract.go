package types

import (
	"encoding/base64"
)

// Creates a new PostMetric contract message from a Metric
func NewMsgExecuteContractPostMetric(metric Metric) MsgExecuteContractPostMetric {
	return MsgExecuteContractPostMetric{
		PostMetric: &MsgPostMetric{
			Key:         metric.Key,
			Value:       metric.Value,
			MetricType:  metric.MetricType,
			UpdateTime:  metric.UpdateTime,
			BlockHeight: metric.BlockHeight,
			Attributes:  base64.StdEncoding.EncodeToString([]byte(metric.Attributes)),
		},
	}
}
