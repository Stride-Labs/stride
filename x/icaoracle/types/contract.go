package types

import (
	"encoding/base64"
)

type MsgInstantiateOracleContract struct {
	AdminAddress string `json:"admin_address"`
}

type postMetric struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	MetricType  string `json:"metric_type"`
	UpdateTime  int64  `json:"update_time"`
	BlockHeight int64  `json:"block_height"`
	Attributes  string `json:"attributes"`
}

type MsgExecuteContractPostMetric struct {
	PostMetric postMetric `json:"post_metric"`
}

func NewMsgExecuteContractPostMetric(metric Metric) MsgExecuteContractPostMetric {
	return MsgExecuteContractPostMetric{
		PostMetric: postMetric{
			Key:         metric.Key,
			Value:       metric.Value,
			MetricType:  metric.MetricType,
			UpdateTime:  metric.UpdateTime,
			BlockHeight: metric.BlockHeight,
			Attributes:  base64.StdEncoding.EncodeToString([]byte(metric.Attributes)),
		},
	}
}
