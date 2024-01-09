#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

echo ">>> Sending usdc tokens to withdrawal ICA to simulate rewards..." 
$NOBLE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(GET_ICA_ADDR $REWARD_CONVERTER_HOST_ZONE withdrawal) 1000000${USDC_DENOM} \
    --from ${NOBLE_VAL_PREFIX}1 -y | TRIM_TX