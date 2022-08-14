#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

mkdir -p $STATE/hermes
mkdir -p $STATE/icq

cp ${SCRIPT_DIR}/config/icq_config.yaml $STATE/icq/config.yaml
cp ${SCRIPT_DIR}/config/hermes_config.toml $STATE/hermes/config.toml

echo "Adding Hermes keys"
$HERMES_CMD keys restore --name rly1 --mnemonic "$HERMES_STRIDE_MNEMONIC" $STRIDE_CHAIN_ID 
$HERMES_CMD keys restore --name rly2 --mnemonic "$HERMES_GAIA_MNEMONIC" $GAIA_CHAIN_ID

echo "Adding ICQ keys"
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride 
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia 
