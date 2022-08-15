#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
docker-compose down
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp

STRIDE_LOGS=$SCRIPT_DIR/logs/stride.log
GAIA_LOGS=$SCRIPT_DIR/logs/gaia.log
HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

# Initialize the state for stride/gaia and relayers
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
sh ${SCRIPT_DIR}/init_relayers.sh

echo "Starting STRIDE chain"
docker-compose up -d stride1 # stride2 stride3 

echo "Starting GAIA chain"
docker-compose up -d gaia1 # gaia2 gaia3

docker-compose logs -f stride1 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $STRIDE_LOGS 2>&1 &
docker-compose logs -f gaia1 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $GAIA_LOGS 2>&1 &

printf "Waiting for STRIDE and GAIA to start..."
( tail -f -n0 $STRIDE_LOGS & ) | grep -q "finalizing commit of block"
( tail -f -n0 $GAIA_LOGS & ) | grep -q "finalizing commit of block"
sleep 5
echo "Done"

printf "Creating connection..."
$HERMES_EXEC create connection $STRIDE_CHAIN_ID $GAIA_CHAIN_ID | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

printf "Creating transfer channel..."
$HERMES_EXEC create channel --port-a transfer --port-b transfer $GAIA_CHAIN_ID connection-0 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Starting relayers"
docker-compose up -d hermes icq

docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 &
docker-compose logs -f icq | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $ICQ_LOGS 2>&1 &


# printf "Creating clients, connections, and transfer channel"
# $RELAYER_EXEC transact link stride-gaia
# echo "DONE"

bash $SCRIPT_DIR/register_host.sh

$SCRIPT_DIR/create_logs.sh &
