#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DEPLOYMENT_NAME="$1"
STRIDE_CHAIN_ID="$2"
NUM_STRIDE_NODES="$3"
STRIDE_ADMIN_MNEMONIC="${@:4}"

echo "Setting up deployment $DEPLOYMENT_NAME"

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh 

echo "Cleaning state"
rm -rf $STATE $STATE/keys.txt

STATE=$SCRIPT_DIR/state
mkdir -p $STATE
touch $STATE/keys.txt

source ${SCRIPT_DIR}/setup_stride_state.sh 
source ${SCRIPT_DIR}/setup_gaia_state.sh 
source ${SCRIPT_DIR}/setup_hermes_state.sh
source ${SCRIPT_DIR}/setup_icq_state.sh 

cp ${SCRIPT_DIR}/install_faucet.sh $STATE/install_faucet.sh

echo "Done"