#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
pkill rly || true
sleep 5
nohup rly start gaia_path -p events --home $SCRIPT_DIR/go-rly >> $RLY_GAIA_LOGS 2>&1 &

