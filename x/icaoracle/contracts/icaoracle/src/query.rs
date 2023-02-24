use cosmwasm_std::{Deps, StdResult, Order};
use crate::state::{CONFIG, METRICS, Config, Metric};
use crate::msg::AllMetricsResponse;

pub fn get_config(deps: Deps) -> StdResult<Config> {
    let config: Config = CONFIG.load(deps.storage)?;
    Ok(config)
}

pub fn get_metric(deps: Deps, key: String) -> StdResult<Metric> {
    let metric: Metric = METRICS.may_load(deps.storage, key)?.unwrap();
    Ok(metric)
} 

pub fn get_all_metrics(deps: Deps) -> StdResult<AllMetricsResponse> {
    let metrics = METRICS
        .range_raw(deps.storage, None, None, Order::Ascending)
        .map(|r| r.map(|(_, v)| v))
        .collect::<StdResult<_>>()?;
        
    Ok(AllMetricsResponse { metrics })
}

