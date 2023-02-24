use cosmwasm_std::{DepsMut, Env, MessageInfo, Response};
use crate::error::ContractError;
use crate::state::{Metric, METRICS, Metadata, CONFIG};

pub fn update_metric(deps: DepsMut, _env: Env, info: MessageInfo, key: String, value: String, metadata: Metadata) -> Result<Response, ContractError> {
    // Only the ICA account can update metrics
    let config = CONFIG.load(deps.storage)?;
    if info.sender != config.admin_address {
        return Err(ContractError::Unauthorized {  });
    }

    // Build the new metric object
    let new_metric = Metric{
        key: key.clone(),
        value,
        metadata,
    };

    // Check if the metric is already in the store
    let old_metric = METRICS.may_load(deps.storage, key.clone())?;

    // If the metric existed already, check that the new metric has a more recent block time
    if let Some(old_metric) = old_metric {
        let old_time = old_metric.metadata.update_time.parse::<i32>().unwrap();
        let new_time = new_metric.metadata.update_time.parse::<i32>().unwrap();
        if new_time < old_time {
            return Err(ContractError::StaleMetric { 
                key: key.clone(), 
                new_time: new_metric.metadata.update_time, 
                old_time: old_metric.metadata.update_time 
            });
        }
    }

    // If the metric does not exist already, or the new one has a more updated time, add it to the store
    METRICS.save(deps.storage, key.clone(), &new_metric)?;
    return Ok(Response::new().add_attribute("action", "update_metric"));
}