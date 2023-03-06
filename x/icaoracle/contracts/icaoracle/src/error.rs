use cosmwasm_std::{StdError};
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("The provided metric (type {metric_type:?}) has invalid metadata attributes")]
    InvalidMetricMetadataAttributes {
        metric_type: String,
    },

    #[error("Invalid denom: {reason}")]
    InvalidDenom {
        reason: String,
    },
}

