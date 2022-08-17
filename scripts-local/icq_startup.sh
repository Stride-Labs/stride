#!/bin/bash 
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/vars.sh
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

printf '%s' "Starting ICQ...               "
nohup $ICQ_CMD run --local >> $ICQ_LOGS 2>&1 &
sleep 5
echo "Done"
