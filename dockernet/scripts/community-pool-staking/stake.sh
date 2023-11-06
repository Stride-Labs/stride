#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

echo ">>> Sending native tokens to deposit ICA to simulate community pool liquid stake..." 
$DYDX_MAIN_CMD tx bank send $(DYDX_ADDRESS) $(GET_ICA_ADDR DYDX community_pool_deposit) 1000000${DYDX_DENOM} --from ${DYDX_VAL_PREFIX}1 -y | TRIM_TX