#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_VAL_PREFIX=$(GET_VAR_VALUE ${HOST_CHAIN}_VAL_PREFIX)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

echo ">>> Transfering native tokens to stride..."
$HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 10000000${HOST_DENOM} \
    --from ${HOST_VAL_PREFIX}1 -y | TRIM_TX
sleep 10

echo ">>> Setting withdrawal address..."
reward_address=$($HOST_MAIN_CMD keys show -a reward)
$HOST_MAIN_CMD tx distribution set-withdraw-addr $reward_address --from delegation -y | TRIM_TX
sleep 10
