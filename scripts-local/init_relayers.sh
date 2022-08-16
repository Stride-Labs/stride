#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Adding Hermes keys"
$HERMES_CMD keys restore --mnemonic "$HERMES_STRIDE_MNEMONIC" $STRIDE_CHAIN
$HERMES_CMD keys restore --mnemonic "$HERMES_GAIA_MNEMONIC" $GAIA_CHAIN

echo "Adding ICQ keys"
mkdir -p $SCRIPT_DIR/state/icq
cp $SCRIPT_DIR/icq/config.yaml $SCRIPT_DIR/state/icq/config.yaml
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride-local
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia-local
