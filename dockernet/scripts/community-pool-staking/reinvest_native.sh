#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_VAL_PREFIX=$(GET_VAR_VALUE ${HOST_CHAIN}_VAL_PREFIX)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Sending native tokens to withdrawal ICA to simulate rewards..." 
$HOST_MAIN_CMD tx bank send ${HOST_VAL_PREFIX}1 $(GET_ICA_ADDR $REWARD_CONVERTER_HOST_ZONE withdrawal) 1000000${HOST_DENOM} \
    --from ${HOST_VAL_PREFIX}1 -y | TRIM_TX