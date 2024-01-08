#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_VAL_PREFIX=$(GET_VAR_VALUE ${HOST_CHAIN}_VAL_PREFIX)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

# Transfer to stride
echo ">>> Transfering native token to Stride..."
$HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000${HOST_DENOM} --from ${HOST_VAL_PREFIX}1 -y | TRIM_TX
sleep 10

#Liquid stake
echo -e "\n>>> Liquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000000 ${HOST_DENOM} --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

# Send stATOM to community pool return address
echo -e "\n>>> Transfer stToken to deposit ICA..."
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(GET_ICA_ADDR $HOST_CHAIN community_pool_deposit) \
    900000st${HOST_DENOM} --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 10
