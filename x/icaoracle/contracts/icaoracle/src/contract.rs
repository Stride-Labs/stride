#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, to_binary};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{Config, CONFIG};
use crate::{execute, query};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:ica-oracle";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let validated_admin_address = deps.api.addr_validate(&msg.admin_address)?;

    let config = Config{
        admin_address: validated_admin_address,
    };

    CONFIG.save(deps.storage, &config)?;

    Ok(Response::new().add_attribute("action", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::PostMetric { 
            key, 
            value, 
            metric_type,
            metadata 
        } => execute::post_metric(deps, info, key, value, metric_type, metadata)
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {  } => to_binary(&query::get_config(deps)?),
        QueryMsg::LatestMetric { key } => to_binary(&query::get_latest_metric(deps, key)?),
        QueryMsg::AllLatestMetrics {  } => to_binary(&query::get_all_latest_metrics(deps)?),
        QueryMsg::HistoricalMetrics { key } => to_binary(&query::get_historical_metrics(deps, key)?),
        QueryMsg::Price { 
            denom, 
            base_denom, 
            params 
        } => to_binary(&query::get_price(deps, denom, base_denom, params)?),
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use crate::contract::{execute, instantiate, query};
    use cosmwasm_std::{attr, Addr, Decimal};
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info, MockStorage, MockApi, MockQuerier};
    use cosmwasm_std::{from_binary, OwnedDeps, DepsMut, Env, MessageInfo, Empty};
    use serde_json::json;

    use crate::msg::{InstantiateMsg, ExecuteMsg, QueryMsg, Metrics, PriceResponse};
    use crate::state::{Metric, Config, Metadata, RedemptionRateAttributes};
    use crate::error::ContractError;

    const ADMIN_ADDRESS: &str = "admin";
    const DENOM: &str = "denom";
    const IBC_DENOM: &str = "ibc/denom";

    // Helper function to setup deps, info, and env in the default test case 
    // In the default case, the message orginates from the ADMIN_ADDRESS
    fn default_setup() -> (OwnedDeps<MockStorage, MockApi, MockQuerier, Empty>, Env, MessageInfo) {
        let env = mock_env();
        let info = mock_info(ADMIN_ADDRESS, &[]);
        let deps = mock_dependencies();
        (deps, env, info)
    }

    // Helper function to instantiate the contract using the default settings
    fn default_instantiate(deps: DepsMut) {
        let (_, env, info) = default_setup();

        let msg = InstantiateMsg{
            admin_address: ADMIN_ADDRESS.to_string(),
        };

        let resp = instantiate(deps, env.clone(), info, msg).unwrap();

        assert_eq!(resp.attributes, vec![
            attr("action", "instantiate")
        ]);
    }

    // Helper function to build a redemption rate object
    // The time field is used for both the time and the block_height
    // It uses a generic denom and ibc/denom
    fn get_test_redemption_rate_metric(key: &str, value: &str, time: u64) -> Metric {
        let redemption_rate_attributes = RedemptionRateAttributes{
            denom: DENOM.to_string(),
            base_denom: IBC_DENOM.to_string(),
        };
        let redemption_rate_attributes = Some(json!(redemption_rate_attributes).to_string());
        let metric = Metric{
            key: key.to_string(),
            value: value.to_string(),
            metric_type: "redemption_rate".to_string(),
            metadata: Metadata { 
                update_time: time,
                block_height: time,
                attributes: redemption_rate_attributes 
            },
        };
        return metric
    }

    // Helper function to build the PostMetric message, given a metric
    fn get_post_metric_msg(metric: Metric) -> ExecuteMsg {
        return ExecuteMsg::PostMetric { 
            key: metric.key.clone(), 
            value: metric.value.clone(),
            metric_type: metric.metric_type.clone(),
            metadata: metric.metadata.clone(),
        };
    }

    #[test]
    fn test_instantiate() {
        let (mut deps, env, _) = default_setup();
        default_instantiate(deps.as_mut());

        // Confirm addresses were set properly
        let msg = QueryMsg::Config {  };
        let resp = query(deps.as_ref(), env, msg).unwrap();
        let config: Config = from_binary(&resp).unwrap();
        assert_eq!(config, Config{ 
            admin_address: Addr::unchecked(ADMIN_ADDRESS.to_string()),
        })
    }

    #[test]
    fn test_post_redemption_rate_metric() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Post a metric
        let metric = get_test_redemption_rate_metric("key1", "1", 1);
        let post_msg = get_post_metric_msg(metric.clone());

        let resp = execute(deps.as_mut(), env.clone(), info, post_msg).unwrap();
        assert_eq!(resp.attributes, vec![
            attr("action", "post_metric")
        ]);

        // Confirm the metric is present in the latest store
        let query_latest_msg = QueryMsg::LatestMetric { 
            key: metric.key.clone(),
        };

        let resp = query(deps.as_ref(), env.clone(), query_latest_msg).unwrap();
        let latest_response: Metric = from_binary(&resp).unwrap();
        assert_eq!(latest_response, metric.clone());

        // Confirm the metric is present in the historical store
        let query_historical_msg = QueryMsg::HistoricalMetrics { 
            key: metric.key.clone(),
        };

        let resp = query(deps.as_ref(), env.clone(), query_historical_msg).unwrap();
        let historical_response: Metrics = from_binary(&resp).unwrap();
        assert_eq!(historical_response, Metrics { metrics: vec![metric.clone()] });
        
        // Confirm the metric was added to the price store
        let query_price_msg = QueryMsg::Price { 
            denom: DENOM.to_string(), 
            base_denom: IBC_DENOM.to_string(), 
            params: None, 
        };
        let resp = query(deps.as_ref(), env.clone(), query_price_msg).unwrap();
        let price_response: PriceResponse = from_binary(&resp).unwrap();
        let expected_price = Decimal::from_str("1").unwrap();
        assert_eq!(price_response, PriceResponse { 
            exchange_rate: expected_price, 
            last_updated: 1 
        });
    }

    #[test]
    fn test_post_different_keys_and_check_latest() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Build three metrics and post messages for each
        let metric1 = get_test_redemption_rate_metric("key1", "1", 1);
        let metric2 = get_test_redemption_rate_metric("key2", "2", 2);
        let metric3 = get_test_redemption_rate_metric("key3", "3", 3);

        let msg1 = get_post_metric_msg(metric1.clone());
        let msg2 = get_post_metric_msg(metric2.clone());
        let msg3 = get_post_metric_msg(metric3.clone());

        // Execute each message 
        execute(deps.as_mut(), env.clone(), info.clone(), msg1).unwrap();
        execute(deps.as_mut(), env.clone(), info.clone(), msg2).unwrap();
        execute(deps.as_mut(), env.clone(), info.clone(), msg3).unwrap();

        // Confirm all metrics are preset and are sorted 
        let msg = QueryMsg::AllLatestMetrics {  };
        let resp = query(deps.as_ref(), env.clone(), msg).unwrap();
        let metric_responses: Metrics = from_binary(&resp).unwrap();
        assert_eq!(metric_responses, Metrics{
            metrics: vec![metric1, metric2, metric3]
        })
    }

    #[test]
    fn test_post_same_key_and_check_historical() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Build four metrics (all with the same key, and with a duplicate value) and post messages for each
        let metric1 = get_test_redemption_rate_metric("key1", "1", 1);
        let metric2 = get_test_redemption_rate_metric("key1", "2", 2);
        let metric3 = get_test_redemption_rate_metric("key1", "3", 2);
        let metric4 = get_test_redemption_rate_metric("key1", "4", 3);

        let msg1 = get_post_metric_msg(metric1.clone());
        let msg2 = get_post_metric_msg(metric2.clone());
        let msg3 = get_post_metric_msg(metric3.clone());
        let msg4 = get_post_metric_msg(metric4.clone());

        // Execute each message out of order, and with msg2 coming before msg3
        execute(deps.as_mut(), env.clone(), info.clone(), msg2).unwrap(); 
        execute(deps.as_mut(), env.clone(), info.clone(), msg1).unwrap();
        execute(deps.as_mut(), env.clone(), info.clone(), msg3).unwrap(); // should get ignored bc duplicate time
        execute(deps.as_mut(), env.clone(), info.clone(), msg4).unwrap();

        // Confirm metrics 1, 2 and 3 are preset and are sorted 
        let msg = QueryMsg::HistoricalMetrics { key: "key1".to_string() };
        let resp = query(deps.as_ref(), env.clone(), msg).unwrap();
        let history_response: Metrics = from_binary(&resp).unwrap();
        assert_eq!(history_response, Metrics{
            metrics: vec![metric1, metric2, metric4]
        })
    }

    #[test]
    fn test_post_metric_unauthorized() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Change info to have non-admin sender
        let mut invalid_info = info.clone();
        invalid_info.sender = Addr::unchecked("not_admin".to_string());

        // Attempt to post the message - it should fail
        let metric = get_test_redemption_rate_metric("key1", "1", 1);
        let post_msg = get_post_metric_msg(metric.clone());
        let resp = execute(deps.as_mut(), env.clone(), invalid_info, post_msg);
        assert_eq!(resp, Err(ContractError::Unauthorized{ }))
    }

    #[test]
    fn test_post_redemption_rate_missing_attributes() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Build a metric object with None for the attributes
        let mut invalid_metric = get_test_redemption_rate_metric("key1", "1", 1);
        invalid_metric.metadata.attributes = None;

        // Attempt to post the message, it should fail
        let post_msg_failure = get_post_metric_msg(invalid_metric.clone());
        let resp = execute(deps.as_mut(), env.clone(), info.clone(), post_msg_failure);
        assert_eq!(resp, Err(ContractError::InvalidMetricMetadataAttributes{ 
            metric_type: invalid_metric.metric_type.clone(),
        }));

        // Now change the metric_type, so that it's not redemption_rate
        let mut valid_metric = invalid_metric.clone();
        valid_metric.metric_type = "something_else".to_string();

        // Now the message should succeed
        let post_msg_success = get_post_metric_msg(valid_metric.clone());
        execute(deps.as_mut(), env.clone(), info, post_msg_success).unwrap();

        // Confirm the metric is present 
        let query_latest_msg = QueryMsg::LatestMetric { 
            key: valid_metric.key.clone(),
        };
        let resp = query(deps.as_ref(), env.clone(), query_latest_msg).unwrap();
        let latest_response: Metric = from_binary(&resp).unwrap();
        assert_eq!(latest_response, valid_metric.clone());
    }

    #[test]
    fn test_post_redemption_rate_invalid_attributes() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Build a metric object with an gibberish string so that it can't be deserialized as attributes
        let mut invalid_metric = get_test_redemption_rate_metric("key1", "1", 1);
        invalid_metric.metadata.attributes = Some("{cantparse}".to_string());

        // Attempt to post the message, it should fail
        let post_msg_failure = get_post_metric_msg(invalid_metric.clone());
        let resp = execute(deps.as_mut(), env.clone(), info.clone(), post_msg_failure);
        assert_eq!(resp, Err(ContractError::InvalidMetricMetadataAttributes{ 
            metric_type: invalid_metric.metric_type.clone(),
        }));

        // Now change the metric_type, so that it's not redemption_rate
        let mut valid_metric = invalid_metric.clone();
        valid_metric.metric_type = "something_else".to_string();

        // Now the message should succeed
        let post_msg_success = get_post_metric_msg(valid_metric.clone());
        execute(deps.as_mut(), env.clone(), info, post_msg_success).unwrap();

        // Confirm the metric is present 
        let query_latest_msg = QueryMsg::LatestMetric { 
            key: valid_metric.key.clone(),
        };
        let resp = query(deps.as_ref(), env.clone(), query_latest_msg).unwrap();
        let latest_response: Metric = from_binary(&resp).unwrap();
        assert_eq!(latest_response, valid_metric.clone());
    }
}
