#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
$HERMES_CMD create connection $STRIDE_CHAIN $GAIA_CHAIN 

echo "Creating transfer channel for Gaia"
$HERMES_CMD create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 

# start go relayer 
nohup rly start gaia_path -p events --home $SCRIPT_DIR/go-rly >> $RLY_GAIA_LOGS 2>&1 &

# echo "Creating connection $STRIDE_CHAIN <> $OSMO_CHAIN"
# $HERMES_CMD create connection $STRIDE_CHAIN $OSMO_CHAIN 

# echo "Creating transfer channel for Osmo"
# $HERMES_CMD create channel --port-a transfer --port-b transfer $OSMO_CHAIN connection-0

# nohup rly start osmo_path -p events --home $SCRIPT_DIR/go-rly >> $RLY_OSMO_LOGS 2>&1 &
# echo "Creating connection $STRIDE_CHAIN <> $JUNO_CHAIN"
# $HERMES_CMD create connection $STRIDE_CHAIN $JUNO_CHAIN 

# echo "Creating transfer channel for Juno"
# $HERMES_CMD create channel --port-a transfer --port-b transfer $JUNO_CHAIN connection-0

