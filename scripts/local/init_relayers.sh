#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Adding Hermes keys"
echo $HERMES_STRIDE_MNEMONIC | $HERMES_CMD keys add -m /dev/stdin $STRIDE_CHAIN
echo $HERMES_GAIA_MNEMONIC | $HERMES_CMD keys add -m /dev/stdin $GAIA_CHAIN

echo "Adding ICQ keys"
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride-local --local
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia-local --local
