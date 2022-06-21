#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# Clear out old keys
rm -rf ~/.hermes/keys

echo "Restoring keys"
echo $RLY_MNEMONIC_1 | $HERMES_CMD keys add -m /dev/stdin $STRIDE_CHAIN
echo $RLY_MNEMONIC_2 | $HERMES_CMD keys add -m /dev/stdin $GAIA_CHAIN

echo "creating hermes identifiers"
$HERMES_CMD tx raw create-client $STRIDE_CHAIN $GAIA_CHAIN 
$HERMES_CMD tx raw conn-init $STRIDE_CHAIN $GAIA_CHAIN 07-tendermint-0 07-tendermint-0 

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
$HERMES_CMD create connection $STRIDE_CHAIN $GAIA_CHAIN 

echo "Creating transfer channel"
$HERMES_CMD create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 > /dev/null
$HERMES_CMD tx raw chan-open-init $STRIDE_CHAIN $GAIA_CHAIN connection-0 transfer transfer > /dev/null

$HERMES_CMD start