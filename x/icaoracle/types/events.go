package types

const (
	EventTypeUpdateOracle    = "update_oracle"
	EventTypeUpdateOracleAck = "update_oracle_ack"

	AttributeKeyOracleChainId     = "oracle_chain_id"
	AttributeKeyMetricID          = "metric_id"
	AttributeKeyMetricKey         = "metric_key"
	AttributeKeyMetricValue       = "metric_value"
	AttributeKeyMetricType        = "metric_type"
	AttributeKeyMetricUpdateTime  = "metric_update_time"
	AttributeKeyMetricBlockHeight = "metric_block_height"
	AttributeKeyMetricAckStatus   = "metric_ack_status"
)
