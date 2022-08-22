#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

HOST_CHAINS="$@"
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log

for host_chain in ${HOST_CHAINS[@]}; do
    echo "Creating connection STRIDE <> $host_chain..." | tee -a $HERMES_LOGS
    $HERMES_EXEC create connection $STRIDE_CHAIN_ID $host_chain | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
    echo "Done"

    echo "Creating transfer channel STRIDE <> $host_chain..." | tee -a $HERMES_LOGS
    $HERMES_EXEC create channel --port-a transfer --port-b transfer $host_chain connection-0 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
    echo "Done"
done

