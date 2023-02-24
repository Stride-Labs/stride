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
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::UpdateMetric { key, value, metadata } => execute::update_metric(deps, env, info, key, value, metadata)
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {  } => to_binary(&query::get_config(deps)?),
        QueryMsg::Metric { key } => to_binary(&query::get_metric(deps, key)?),
        QueryMsg::AllMetrics { } => to_binary(&query::get_all_metrics(deps)?),
    }
}

#[cfg(test)]
mod tests {
    use crate::ContractError;
    use crate::contract::{execute, instantiate, query};
    use cosmwasm_std::{attr, Addr};
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info, MockStorage, MockApi, MockQuerier};
    use cosmwasm_std::{from_binary, OwnedDeps, DepsMut, Env, MessageInfo, Empty};

    use crate::msg::{InstantiateMsg, ExecuteMsg, QueryMsg, AllMetricsResponse};
    use crate::state::{Metric, Config, Metadata};

    const ADMIN_ADDRESS: &str = "admin";

    fn default_setup() -> (OwnedDeps<MockStorage, MockApi, MockQuerier, Empty>, Env, MessageInfo) {
        let env = mock_env();
        let info = mock_info(ADMIN_ADDRESS, &[]);
        let deps = mock_dependencies();
        (deps, env, info)
    }

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
    fn test_update_metric() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Add a metric
        let metric = Metric{
            key: "key".to_string(), 
            value: "value".to_string(),
            metadata: Metadata{
                update_height: "1".to_string(),
                update_time: "2".to_string(),
            }
        };
        let msg = ExecuteMsg::UpdateMetric { 
            key: metric.key.clone(), 
            value: metric.value.clone(),
            metadata: metric.metadata.clone(),
        };

        let resp = execute(deps.as_mut(), env.clone(), info, msg).unwrap();
        assert_eq!(resp.attributes, vec![
            attr("action", "update_metric")
        ]);

        // Confirm the metric is present
        let msg = QueryMsg::Metric { 
            key: "key".to_string(),
        };

        let resp = query(deps.as_ref(), env.clone(), msg).unwrap();
        let metric_response: Metric = from_binary(&resp).unwrap();
        assert_eq!(metric_response, metric.clone());
    }

    #[test]
    fn test_update_stale_metric() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        let old_time = "10".to_string();
        let new_time = "1".to_string();

        // Add a metric
        let metric = Metric{
            key: "key".to_string(), 
            value: "value".to_string(),
            metadata: Metadata{
                update_height: "1".to_string(),
                update_time: old_time.clone(),
            }
        };
        let msg = ExecuteMsg::UpdateMetric { 
            key: metric.key.clone(), 
            value: metric.value.clone(),
            metadata: metric.metadata.clone(),
        };

        let resp = execute(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();
        assert_eq!(resp.attributes, vec![
            attr("action", "update_metric")
        ]);

        // Confirm the metric is present
        let msg = QueryMsg::Metric { 
            key: "key".to_string(),
        };

        let resp = query(deps.as_ref(), env.clone(), msg).unwrap();
        let metric_response: Metric = from_binary(&resp).unwrap();
        assert_eq!(metric_response, metric.clone());

        // Attempt to add another metric that has an older timestamp
        let metric = Metric{
            key: "key".to_string(), 
            value: "value".to_string(),
            metadata: Metadata{
                update_height: "1".to_string(),
                update_time: new_time.clone(),
            }
        };
        let msg = ExecuteMsg::UpdateMetric { 
            key: metric.key.clone(), 
            value: metric.value.clone(),
            metadata: metric.metadata.clone(),
        };

        let resp = execute(deps.as_mut(), env.clone(), info, msg).unwrap_err();
        assert_eq!(resp, ContractError::StaleMetric { 
            key: metric.key.clone(), 
            new_time: new_time.clone(), 
            old_time: old_time.clone(),
        })
    }

    #[test]
    fn test_get_all_metrics() {
        // Instantiate contract
        let (mut deps, env, info) = default_setup();
        default_instantiate(deps.as_mut());

        // Add two metrics
        let metric1 = Metric{
            key: "key1".to_string(), 
            value: "value1".to_string(),
            metadata: Metadata{update_time: "10".to_string(), update_height: "1".to_string()},
        };
        let msg = ExecuteMsg::UpdateMetric {
            key: metric1.key.clone(),
            value: metric1.value.clone(),
            metadata: metric1.metadata.clone(),
        };
        execute(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        let metric2 = Metric{
            key: "key2".to_string(), 
            value: "value2".to_string(),
            metadata: Metadata{update_time: "20".to_string(), update_height: "2".to_string()},
        };
        let msg = ExecuteMsg::UpdateMetric { 
            key: metric2.key.clone(),
            value: metric2.value.clone(),
            metadata: metric2.metadata.clone(),
         };
        execute(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Confirm all metrics are preset
        let msg = QueryMsg::AllMetrics { };
        let resp = query(deps.as_ref(), env.clone(), msg).unwrap();
        let metric_responses: AllMetricsResponse = from_binary(&resp).unwrap();
        assert_eq!(metric_responses, AllMetricsResponse{
            metrics: vec![metric1, metric2]
        })
    }
}
