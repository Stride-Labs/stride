use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin_address: Addr,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Metric {
    pub key: String,
    pub value: String,
    pub metadata: Metadata
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Metadata {
    pub update_time: String,
    pub update_height: String,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const METRICS: Map<String, Metric> = Map::new("metrics");