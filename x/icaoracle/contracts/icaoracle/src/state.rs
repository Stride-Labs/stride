use std::{collections::VecDeque, usize};
use cosmwasm_std::{Addr, Decimal};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use strum_macros::EnumString;

use cw_storage_plus::{Item, Map};


// The contract config consists of an admin address 
// The admin address will be the ICA address for the account that's 
//  owned by the source chain and lives on the contract chain
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin_address: Addr,
}

// This contract represents a generic key value store 
// A "metric" is the term for a piece of information stored
// Each metric has a higher level category that helps inform if any other,
// metric-specific logic needs to be run
// i.e. For redemption rates, there is an expected format for the attributes 
// field with additional metadata
#[derive(EnumString)]
pub enum MetricType {
  #[strum(serialize = "redemption_rate")]
  RedemptionRate,
}

// The Metric struct represents the base unit for the generic oracle key-value store
// The key/value represent the main piece of data that is intended to be stored
// The metric_type represents a high level category for the metric
// The metadata field contains any additional info such as the block height/time
//  as well as any other additional context that's needed
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Metric {
    pub key: String,
    pub value: String,
    pub metric_type: String,
    pub metadata: Metadata
}

// QUESTION: What's the best way to handle these attributes, considering
// there's a different schema for each message type?
//
// The Metadata struct stores additional context for each metric  
// The update_time/block_height are the time and block height at which 
//   the value was updated on the source chain
// The attributes field is optional for any additional metric-specific context
// e.g. For redemption_rate metrics, the IBC Denom is stored as an attribute
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Metadata {
    pub update_time: u64,
    pub block_height: u64,
    pub attributes: Option<String>
}

// For use in price oracles, the RedemptionRate metric requires the stToken denom
//  and the native token's IBC Denom, in order to align with the expected price query
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct RedemptionRateAttributes {
    pub denom: String,
    pub base_denom: String,
}

// The Price struct represents the exchange rate of a denom/base_denom pair
// In the case of the redemption_rate of an stToken, the denom would be the 
//  stToken denom, and the base_denom would be the ibc hash of the native token
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Price {
    pub denom: String,
    pub base_denom: String,
    pub exchange_rate: Decimal,
    pub last_updated: u64,
}

// Helper function to get the key for the Price store which is built from the denom and base_denom
pub fn get_price_key(denom: &str, base_denom: &str) -> String {
    return format!("{}-{}", denom, base_denom)
}

// The history of each metric is also stored in the contract to enable 
//   historical queries or averaging/smoothing 
// For each metric, the history is stored in a deque with a max capacity
// The deque is sorted by the time at which the metric was updated on the source chain
// This allows for the efficient insertion of new metrics (to the front of the deque in most cases)
//  as well as the efficient range look of the most recent items (also pulled from the front of the deque)
// Since there is no current use case for storing all history indefinitely, the list is pruned by removing
//  elements from the back of the deque when the capacity has been reached
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct MetricHistory {
    deque: VecDeque<Metric>,
    capacity: u64,
}
impl MetricHistory {
    // Instantiates a new deque with a fixed capacity
    pub fn new() -> Self {
        let capacity = 100;
        MetricHistory {
            deque: VecDeque::with_capacity(capacity as usize),
            capacity,
        }
    }

    // Adds a new metric to the deque
    // It first checks if a metric with the same timestamp is found
    //   If a metric is found with the same timestamp
    //     -> that implies the metric is a duplciate
    //     -> binary_search_by_key will return Ok
    //     -> leaves the old metric in the store
    //   If the same timestamp is not found 
    //      -> that implies this metric is new
    //      -> binary_search_by_key will return Err
    //      -> we insert the new metric to the list
    // New items are added such that the deque remains ordered by timestamp
    // Old items are removed from the back of the deque when capacity is reached
    pub fn add(&mut self, metric: Metric) {
        if let Err(index) = self.deque.binary_search_by_key(&metric.metadata.update_time, |m| m.metadata.update_time) {
            self.deque.insert(index, metric);
            if self.deque.len() > self.capacity as usize {
                self.deque.pop_back();
            }
        }
    }

    // Grabs the most recent metric from the deque
    pub fn get_latest(&self) -> Option<Metric> {
        self.deque.front().cloned()
    }

    // Grabs the most recent N metrics from the deque
    pub fn get_latest_range(&self, n: usize) -> Vec<Metric> {
        self.deque.iter().take(n).cloned().collect()
    }

    // Returns all metrics as a list
    pub fn get_all_metrics(&self) -> Vec<Metric> {
        self.deque.iter().cloned().collect()
    }
}


// The CONFIG store stores contract configuration such as the admin address 
pub const CONFIG: Item<Config> = Item::new("config");

// The LATEST_METRIC store stores the most recent update for each metric
// It is key'd by the metric's "key" field
pub const LATEST_METRICS: Map<&str, Metric> = Map::new("latest_metrics");

// The HISTORICAL_METRICS store stores the full history of a metric 
// It is key'd on the metric "key" field, but consists of a list (deque) of each metric sorted by update time
pub const HISTORICAL_METRICS: Map<&str, MetricHistory> = Map::new("historical_metrics");

// The PRICES store is dedicated to pricing-specific metrics and facilitate queries from price oracles
// It is key'd on the "denom" + "base_denom" fields
pub const PRICES: Map<&str, Price> = Map::new("prices");

