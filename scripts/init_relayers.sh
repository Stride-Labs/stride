#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

mkdir -p $STATE/hermes
mkdir -p $STATE/icq
mkdir -p $STATE/relayer/config

cp ${SCRIPT_DIR}/config/icq_config.yaml $STATE/icq/config.yaml
cp ${SCRIPT_DIR}/config/hermes_config.toml $STATE/hermes/config.toml
cp ${SCRIPT_DIR}/config/relayer_config.yaml $STATE/relayer/config/config.yaml

echo "Adding Hermes keys"
$HERMES_CMD keys restore --name rly1 --mnemonic "$HERMES_STRIDE_MNEMONIC" $STRIDE_CHAIN_ID 
$HERMES_CMD keys restore --name rly2 --mnemonic "$HERMES_GAIA_MNEMONIC" $GAIA_CHAIN_ID

# echo "Adding Relayer keys"
# $RELAYER_CMD keys restore stride rly1 "$RELAYER_STRIDE_MNEMONIC"
# $RELAYER_CMD keys restore gaia rly2 "$RELAYER_GAIA_MNEMONIC" 
# $RELAYER_CMD paths new $STRIDE_CHAIN_ID $GAIA_CHAIN_ID stride-gaia

echo "Adding ICQ keys"
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore icq1 --chain stride 
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore icq2 --chain gaia 
