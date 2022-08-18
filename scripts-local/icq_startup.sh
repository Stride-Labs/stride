#!/bin/bash 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh

while true; do
    printf '%s' "(Re)starting ICQ (hot-restart enabled)..."
    $ICQ_CMD run --local >> $ICQ_LOGS 2>&1 &
    WAIT_FOR_STRING $ICQ_LOGS 'panic:'
done
sleep 5
echo "Done"
