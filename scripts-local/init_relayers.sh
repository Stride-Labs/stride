#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Adding Hermes keys"
TMP_MNEMONICS=${SCRIPT_DIR}/state/mnemonic.txt 

echo "$HERMES_STRIDE_MNEMONIC" > mnemonic.txt 
$HERMES_CMD keys add --key-name rly1 --chain $STRIDE_CHAIN_ID --mnemonic-file $TMP_MNEMONICS --overwrite >> $KEYS_LOGS 2>&1
echo "$HERMES_GAIA_MNEMONIC" > mnemonic.txt 
$HERMES_CMD keys add --key-name rly2 --chain $GAIA_CHAIN --mnemonic-file $TMP_MNEMONICS --overwrite >> $KEYS_LOGS 2>&1
echo "$HERMES_JUNO_MNEMONIC" > mnemonic.txt 
$HERMES_CMD keys add --key-name rly3 --chain $JUNO_CHAIN --mnemonic-file $TMP_MNEMONICS --overwrite >> $KEYS_LOGS 2>&1
echo "$HERMES_OSMO_MNEMONIC" > mnemonic.txt 
$HERMES_CMD keys add --key-name rly4 --chain $OSMO_CHAIN --mnemonic-file $TMP_MNEMONICS --overwrite >> $KEYS_LOGS 2>&1
rm -f $TMP_MNEMONICS

echo "Adding ICQ keys"

mkdir -p $SCRIPT_DIR/state/icq
cp $SCRIPT_DIR/icq/config.yaml $SCRIPT_DIR/state/icq/config.yaml
echo $ICQ_STRIDE_MNEMONIC | $ICQ_CMD keys restore stridekey --chain stride-local --local >> $KEYS_LOGS 2>&1
echo $ICQ_GAIA_MNEMONIC | $ICQ_CMD keys restore gaiakey --chain gaia-local --local >> $KEYS_LOGS 2>&1
echo $ICQ_JUNO_MNEMONIC | $ICQ_CMD keys restore junokey --chain juno-local --local >> $KEYS_LOGS 2>&1
echo $ICQ_OSMO_MNEMONIC | $ICQ_CMD keys restore osmokey --chain osmo-local --local >> $KEYS_LOGS 2>&1

# rly keys restore gaia gaiarly "$RLY_GAIA_MNEMONIC"
# rly keys restore stride striderly "$RLY_STRIDE_MNEMONIC"

