#!/bin/bash

set -eu
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

source ${SCRIPT_DIR}/vars.sh

while true; do
    # transactions logs
    TEMP_LOGS_DIR=$SCRIPT_DIR/logs/temp
    mkdir -p $TEMP_LOGS_DIR

    $STRIDE_CMD q txs --events message.module=interchainquery --limit=100000 > $TEMP_LOGS_DIR/stakeibc-events.log
    $STRIDE_CMD q txs --events message.module=stakeibc --limit=100000 > $TEMP_LOGS_DIR/icq-events.log

    mv $TEMP_LOGS_DIR/*.log $SCRIPT_DIR/logs/
    sleep 3
done 