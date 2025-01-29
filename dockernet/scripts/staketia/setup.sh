#!/bin/bash
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN="${HOST_CHAINS[0]}"
HOST_MAIN_CMD=$(GET_VAR_VALUE   ${HOST_CHAIN}_MAIN_CMD)
HOST_VAL_PREFIX=$(GET_VAR_VALUE ${HOST_CHAIN}_VAL_PREFIX)
HOST_DENOM=$(GET_VAR_VALUE      ${HOST_CHAIN}_DENOM)

${HOST_MAIN_CMD} tx bank send gval1 cosmos1n4reqytr7arvpk5z2ute274h2yukcss8dtxjyd 1000000000uatom --from gval1 -y --fees 1000000ufee
sleep 3

echo ">>> Transfering native tokens to stride..."
$HOST_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 10000000${HOST_DENOM} \
    --from ${HOST_VAL_PREFIX}1 -y --fees 1000000ufee | TRIM_TX
sleep 10

# echo ">>> Setting withdrawal address..."
# reward_address=$($HOST_MAIN_CMD keys show -a reward)
# $HOST_MAIN_CMD tx distribution set-withdraw-addr $reward_address --from delegation -y --fees 1000000ufee | TRIM_TX
# sleep 10
