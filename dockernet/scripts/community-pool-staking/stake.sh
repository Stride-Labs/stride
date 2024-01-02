#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_VAL_PREFIX=$(GET_VAR_VALUE ${HOST_CHAIN}_VAL_PREFIX)
HOST_VAL_ADDRESS=$(${HOST_CHAIN}_ADDRESS)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Sending native tokens to deposit ICA to simulate community pool liquid stake..." 
$HOST_MAIN_CMD tx bank send $HOST_VAL_ADDRESS $(GET_ICA_ADDR $HOST_CHAIN community_pool_deposit) 1000000${HOST_DENOM} --from ${HOST_VAL_PREFIX}1 -y | TRIM_TX