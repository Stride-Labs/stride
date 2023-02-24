use cosmwasm_schema::{cw_serde, QueryResponses};
use crate::state::{Config, Metric, Metadata};

#[cw_serde]
pub struct InstantiateMsg {
    pub admin_address: String,
}

// QUESTION: Should this use the Metric type instead? 
// It's a little cleaner but also leads to an unnecessary level of nesting in the message
// { "update_metric": { "metric": {"key": ... }}}
#[cw_serde]
pub enum ExecuteMsg {
    UpdateMetric {
        key: String,
        value: String,
        metadata: Metadata,
    }
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Config)]
    Config {},

    #[returns(Metric)]
    Metric {
        key: String,
    },

    #[returns(AllMetricsResponse)]
    AllMetrics {},
}

#[cw_serde]
pub struct AllMetricsResponse {
    pub metrics: Vec<Metric>,
}

