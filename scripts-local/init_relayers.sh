#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Adding Hermes keys"
$HERMES_CMD keys restore --mnemonic "$HERMES_STRIDE_MNEMONIC" $STRIDE_CHAIN
$HERMES_CMD keys restore --mnemonic "$HERMES_GAIA_MNEMONIC" $GAIA_CHAIN
$HERMES_CMD keys restore --mnemonic "$HERMES_JUNO_MNEMONIC" $JUNO_CHAIN

echo "Adding ICQ keys"
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride-local --local
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia-local --local
echo $ICQ_JUNO_MNEMONIC | $ICQ_CMD keys restore junokey --chain juno-local --local
