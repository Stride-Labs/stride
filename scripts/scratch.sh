#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID