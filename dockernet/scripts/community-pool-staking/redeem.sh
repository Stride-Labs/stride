#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# Transfer to stride
echo ">>> Transfering native token to Stride..."
$DYDX_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000${DYDX_DENOM} --from ${DYDX_VAL_PREFIX}1 -y | TRIM_TX
sleep 10

#Liquid stake
echo -e "\n>>> Liquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000000 ${DYDX_DENOM} --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 5

# Send stATOM to community pool return address
echo -e "\n>>> Transfer stToken to deposit ICA..."
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(GET_ICA_ADDR DYDX community_pool_deposit) \
    900000${STDYDX_DENOM} --from ${STRIDE_VAL_PREFIX}1 -y | TRIM_TX
sleep 10
