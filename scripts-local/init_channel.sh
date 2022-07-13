#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
$HERMES_CMD create connection $STRIDE_CHAIN $GAIA_CHAIN 

echo "Creating transfer channel"
$HERMES_CMD create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 
