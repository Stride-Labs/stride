use cosmwasm_std::{DepsMut, MessageInfo, Response, Decimal};
use std::str::FromStr;
use crate::error::ContractError;
use crate::helpers::validate_native_denom;
use crate::state::{
    Metric, MetricHistory, MetricType, Metadata, RedemptionRateAttributes, Price, 
    LATEST_METRICS, HISTORICAL_METRICS, PRICES, CONFIG,
    get_price_key,
};

// QUESTION: Judging from the unit tests, it looks like this function is not atomic
// If so, do we have to revert store writes to the LATEST_METRIC and HISTORICAL_METRIC store if the subsequent logic fails

// Stores a given metric passed via an ICA from a source chain
// The oracle stores each metric generically with key and value attributes
//
// The metric is stored in the LATEST_STORE if either:
//   * a metric with that key has never been added, OR
//   * a metric with that key has been added, 
//       but the previously added metric has an older timestamp than the current metric
//
// The metric is added to the HISTORICAL store if:
//   * a metric with that key and time combo has never been submitted, AND
//   * the historical list is not at capacity, OR
//      * the historical list is at capacity, but the metric is more recent than the oldest metric in the store
//
// Only metrics with metric_type "redemption_rate" are added to the PRICES store
// If the metric is of type "redemption_rate", it is added to the PRICES store if:
//   * a price with the denom + base_denom pair has never been added, OR
//   * a price with the denom + base_denom pair has been added, 
//       but the previously added price has an older timestamp than the current price, AND
//     * valid metadata attributes are supplied with denom and base_denom fields, each with a valid native denom
pub fn post_metric(
    deps: DepsMut, 
    info: MessageInfo, 
    key: String, 
    value: String, 
    metric_type: String, 
    metadata: Metadata
) -> Result<Response, ContractError> {
    // Only the ICA account can post metrics
    let config = CONFIG.load(deps.storage)?;
    if info.sender != config.admin_address {
        return Err(ContractError::Unauthorized {  });
    }

    // Build the new metric object
    let new_metric = Metric{
        key: key.clone(),
        value: value.clone(),
        metric_type: metric_type.clone(),
        metadata: metadata.clone(),
    };

    // Only save the new metric to the latest store if it doesn't already exist, 
    // or if it's more recent than the existing metric
    let save_to_latest_store: bool = match LATEST_METRICS.may_load(deps.storage, &key)? {
        // If there's an existing metric, only store the new one if it's time is more recent
        Some(old_metric) => new_metric.metadata.update_time > old_metric.metadata.update_time,
        None => true, // Add the new metric if it doesn't exist
    };
    if save_to_latest_store {
        LATEST_METRICS.save(deps.storage, &key, &new_metric)?;
    }

    // Add the metric to the historical store
    let mut history = match HISTORICAL_METRICS.may_load(deps.storage, &key)? {
        Some(history) => history,
        None => MetricHistory::new(),
    };
    history.add(new_metric.clone());
    HISTORICAL_METRICS.save(deps.storage, &key, &history)?;

    // Parse the metric_type field and handle any other metric-type specific cases
    match MetricType::from_str(&metric_type) {
        Ok(metric_type) => {
            match metric_type {
                // If the metric is a redemption rate update, add a price record to the store
                MetricType::RedemptionRate => {
                    // Deserialize a the metric attributes to get the denom and base denom
                    let attributes: RedemptionRateAttributes = match metadata.attributes.as_ref() {
                        Some(attributes) => {
                            serde_json::from_str(&attributes).map_err(|_| ContractError::InvalidMetricMetadataAttributes {
                                metric_type: new_metric.metric_type.clone(),
                            })?
                        },
                        None => return Err(ContractError::InvalidMetricMetadataAttributes { metric_type: new_metric.metric_type.clone() })
                    };
                    let denom = attributes.denom.clone();
                    let base_denom = attributes.base_denom.clone();

                    // Validate each denom
                    validate_native_denom(&denom)?;
                    validate_native_denom(&base_denom)?;

                    // Store the price in the prices table
                    let exchange_rate = Decimal::from_str(&new_metric.value)?;
                    let new_price = Price{
                        denom: denom.clone(),
                        base_denom: base_denom.clone(),
                        exchange_rate,
                        last_updated: new_metric.metadata.update_time,
                    };

                    let price_key = get_price_key(&denom, &base_denom);

                    let save_to_price_store: bool = match PRICES.may_load(deps.storage, &price_key)? {
                        // If there's an existing price, only store the new one if it's time is more recent
                        Some(old_price) => new_price.last_updated > old_price.last_updated,
                        None => true, // Add the new metric if it doesn't exist
                    };
                    if save_to_price_store {
                        PRICES.save(deps.storage, &price_key, &new_price)?;
                    }
                },
            }
        }
        Err(_) => {}
    }

    return Ok(Response::new().add_attribute("action", "post_metric"));
}
