#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Creating clients"
$HERMES_CMD tx raw create-client $STRIDE_CHAIN $GAIA_CHAIN 
$HERMES_CMD tx raw conn-init $STRIDE_CHAIN $GAIA_CHAIN 07-tendermint-0 07-tendermint-0 

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
$HERMES_CMD create connection $STRIDE_CHAIN $GAIA_CHAIN 

echo "Creating transfer channel"
$HERMES_CMD create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 
$HERMES_CMD tx raw chan-open-init $STRIDE_CHAIN $GAIA_CHAIN connection-0 transfer transfer 

