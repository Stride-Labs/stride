#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh
source ${SCRIPT_DIR}/denoms.sh

LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 1) # channel-0, validator 2, recordId 1
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $LSM_TOKEN_DENOM --from staker1 --gas auto -y | TRIM_TX

sleep 5
LSM_TOKEN_DENOM=$(GET_LSM_IBC_TOKEN_DENOM 0 2 2) # channel-0, validator 2, recordId 2
$STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake 1000000 $LSM_TOKEN_DENOM --from staker2 --gas auto -y | TRIM_TX