#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_CHAIN_ID=$(GET_VAR_VALUE   ${HOST_CHAIN}_CHAIN_ID)

$STRIDE_MAIN_CMD tx stakeibc set-rebate $HOST_CHAIN_ID 100000 --from ${STRIDE_VAL_PREFIX}1 -y