#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Adding Hermes keys"
$HERMES_CMD keys restore --mnemonic "$HERMES_STRIDE_MNEMONIC" $STRIDE_CHAIN >> $KEYS_LOGS 2>&1 
$HERMES_CMD keys restore --mnemonic "$HERMES_GAIA_MNEMONIC" $GAIA_CHAIN >> $KEYS_LOGS 2>&1  
# $HERMES_CMD keys restore --mnemonic "$HERMES_JUNO_MNEMONIC" $JUNO_CHAIN >> $KEYS_LOGS 2>&1 
# $HERMES_CMD keys restore --mnemonic "$HERMES_OSMO_MNEMONIC" $OSMO_CHAIN >> $KEYS_LOGS 2>&1 

echo "Adding ICQ keys"
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride-local --local >> $KEYS_LOGS 2>&1 
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia-local --local >> $KEYS_LOGS 2>&1 
# echo $ICQ_JUNO_MNEMONIC | $ICQ_CMD keys restore junokey --chain juno-local --local >> $KEYS_LOGS 2>&1 
# echo $ICQ_OSMO_MNEMONIC | $ICQ_CMD keys restore osmokey --chain osmo-local --local >> $KEYS_LOGS 2>&1 

# rly keys restore gaia gaiarly "$RLY_GAIA_MNEMONIC"
# rly keys restore stride striderly "$RLY_STRIDE_MNEMONIC"