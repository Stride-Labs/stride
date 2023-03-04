use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, Decimal};
use crate::state::{Config, Metric, Metadata};
#[cw_serde]
pub struct InstantiateMsg {
    pub admin_address: String,
}

// QUESTION: Should this use the Metric type instead? 
// It's a little cleaner but also leads to an unnecessary level of nesting in the message
// { "post_metric": { "metric": {"key": ... }}}
#[cw_serde]
pub enum ExecuteMsg {
    PostMetric {
        key: String,
        value: String,
        metric_type: String,
        metadata: Metadata,
    }
}

// Options to query the following:
//  - The config
//  - The most updated metric given the metric's key
//  - All the the most updated metrics
//  - The full history of a metric given a key
//  - The price of a given token - used for Price oracles 
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Config)]
    Config {},

    #[returns(Metric)]
    LatestMetric {
        key: String,
    },

    #[returns(Metrics)]
    AllLatestMetrics {},

    #[returns(Metrics)]
    HistoricalMetrics {
        key: String,
    },

    #[returns(PriceResponse)]
    Price {
        denom: String, 
        base_denom: String,
        params: Option<Binary>,
    },
}

#[cw_serde]
pub struct Metrics {
    pub metrics: Vec<Metric>,
}

#[cw_serde]
pub struct PriceResponse {
    pub exchange_rate: Decimal, 
    pub last_updated: u64, 
}
