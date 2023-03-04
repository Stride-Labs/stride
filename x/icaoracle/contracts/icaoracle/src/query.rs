use cosmwasm_std::{Deps, StdResult, Order, Binary, StdError};
use crate::state::{Config, Metric, Price, CONFIG, LATEST_METRICS, PRICES, HISTORICAL_METRICS, get_price_key};
use crate::msg::{Metrics, PriceResponse};

// Returns the contract configuration
pub fn get_config(deps: Deps) -> StdResult<Config> {
    let config: Config = CONFIG.load(deps.storage)?;
    Ok(config)
}

// Returns the most up-to-date metric, given the metric's key
pub fn get_latest_metric(deps: Deps, key: String) -> StdResult<Metric> {
    let metric = LATEST_METRICS.load(deps.storage, &key)?;
    Ok(metric)
} 

// Returns the most up-to-date metric for all metrics stored
pub fn get_all_latest_metrics(deps: Deps) -> StdResult<Metrics> {
    let metrics = LATEST_METRICS
        .range_raw(deps.storage, None, None, Order::Ascending)
        .map(|r| r.map(|(_, v)| v))
        .collect::<StdResult<_>>()?;
        
    Ok(Metrics { metrics })
}

// Returns the full history of a given metric, sorted by the time at which it was updated
pub fn get_historical_metrics(deps: Deps, key: String) -> StdResult<Metrics> {
    let metrics_history = HISTORICAL_METRICS.load(deps.storage, &key)?;
    let metrics = metrics_history.get_all_metrics();
    Ok(Metrics { metrics })
}

// Returns the price of a given token and the time that it was last updated (used for price oracles)
pub fn get_price(deps: Deps, denom: String, base_denom: String, params: Option<Binary>) -> StdResult<PriceResponse> {
    // The params field of the price query should always be None
    if let Some(_) = params {
        return Err(StdError::generic_err("invalid query request - params must be None"))
    }

    // Grab the prices by looking it up by denom
    // Then confirm that the base_denom of the price object lines up with the base_denom from the query 
    let price_key = get_price_key(&denom, &base_denom);
    let price: Price = PRICES.load(deps.storage, &price_key)?;
    if price.base_denom != base_denom {
        return Err(StdError::generic_err(
            format!("invalid query request - base_denom ({}) does not match denom ({}) - expected: {}", 
            base_denom, denom, price.base_denom)
        ))
    };

    Ok(PriceResponse{ exchange_rate: price.exchange_rate, last_updated: price.last_updated })
}
