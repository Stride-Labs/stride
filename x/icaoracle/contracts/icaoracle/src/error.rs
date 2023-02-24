use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},
    
    #[error("The provided metric update with time ({new_time:?}) is stale. Metric for key ({key:?}) already exists with newer timestamp ({old_time:?})")]
    StaleMetric {
        key: String,
        new_time: String, 
        old_time: String,
    },
}
