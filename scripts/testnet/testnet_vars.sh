#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "$1" ]
then
    echo "Error, you must pass deployment name in. E.g. \"sh setup_testnet_state.sh internal\""
    exit 1
fi

DEPLOYMENT_NAME="$1"
echo "Setting up deployment $DEPLOYMENT_NAME"

STATE=$SCRIPT_DIR/state
PORT_ID=26656
DOMAIN=stridenet.co

BLOCK_TIME=5s

STRIDE_CMD="$SCRIPT_DIR/../../build/strided"
STRIDE_MAIN_ENDPOINT="stride-node1.${DEPLOYMENT_NAME}.${DOMAIN}"

GAIA_CMD="$SCRIPT_DIR/../../build/gaiad --home $STATE/gaia"
GAIA_MAIN_ENDPOINT="gaia.${DEPLOYMENT_NAME}.${DOMAIN}"

HERMES_STRIDE_ACCT=rly1
HERMES_GAIA_ACCT=rly2
HERMES_CMD="$SCRIPT_DIR/../../build/hermes/release/hermes -c $SCRIPT_DIR/hermes/config.toml"

ICQ_STRIDE_ACCT=icq1
ICQ_GAIA_ACCT=icq2
ICQ_CMD="$SCRIPT_DIR/../../build/interchain-queries --home $STATE/icq"

GET_MNEMONIC() {
  grep -i -A 10 "\- name: $1" "$STATE/keys.txt" | tail -n 1
}

GET_ADDRESS() {
  grep -i -A 10 "\- name: val1" scripts/testnet/state/keys.txt | sed -n 3p | awk '{printf $2}'
}
