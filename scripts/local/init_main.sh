#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE ./icq/keys

# TODO(TEST-117) Modularize/generalize chain init scripts 
# Initialize the state for stride/gaia and relayers
# ignite chain init
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
# sh ${SCRIPT_DIR}/init_hermes.sh
# sh ${SCRIPT_DIR}/init_icq.sh

# Register host zone
# ICA staking test
# first register host zone for ATOM chain
# TODO(TEST-118) Improve integration test timing

# ATOM='uatom'
# IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
# CSLEEP 60
# docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided tx stakeibc register-host-zone connection-0 $ATOM $IBCATOM channel-0 --chain-id $STRIDE_CHAIN --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
