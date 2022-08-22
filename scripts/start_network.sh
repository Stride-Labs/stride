#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
docker-compose down
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp
mkdir -p $SCRIPT_DIR/logs

HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

# Initialize the state for stride/gaia and relayers
sh ${SCRIPT_DIR}/init_chain.sh STRIDE
sh ${SCRIPT_DIR}/init_chain.sh GAIA
sh ${SCRIPT_DIR}/init_chain.sh JUNO
sh ${SCRIPT_DIR}/init_chain.sh OSMO

sh ${SCRIPT_DIR}/start_chain.sh STRIDE GAIA JUNO OSMO
sh ${SCRIPT_DIR}/init_relayers.sh STRIDE GAIA JUNO OSMO

echo "Creating connection STRIDE <> GAIA..."
$HERMES_EXEC create connection $STRIDE_CHAIN_ID $GAIA_CHAIN_ID | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Creating transfer channel STRIDE <> GAIA..."
$HERMES_EXEC create channel --port-a transfer --port-b transfer $GAIA_CHAIN_ID connection-0 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Creating connection STRIDE <> JUNO..."
$HERMES_EXEC create connection $STRIDE_CHAIN_ID $JUNO_CHAIN_ID | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Creating transfer channel STRIDE <> JUNO..."
$HERMES_EXEC create channel --port-a transfer --port-b transfer $JUNO_CHAIN_ID connection-0 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Creating connection STRIDE <> OSMO..."
$HERMES_EXEC create connection $STRIDE_CHAIN_ID $OSMO_CHAIN_ID | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

echo "Creating transfer channel STRIDE <> OSMO..."
$HERMES_EXEC create channel --port-a transfer --port-b transfer $OSMO_CHAIN_ID connection-0 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 
echo "Done"

# printf "Creating clients, connections, and transfer channel"
# $RELAYER_EXEC transact link stride-gaia
# echo "DONE"

echo "Starting relayers"
docker-compose up -d hermes icq

docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 &
docker-compose logs -f icq | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $ICQ_LOGS 2>&1 &

( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "Hermes has started"

bash $SCRIPT_DIR/register_host.sh GAIA connection-0 channel-0
bash $SCRIPT_DIR/register_host.sh JUNO connection-1 channel-1
bash $SCRIPT_DIR/register_host.sh OSMO connection-2 channel-2

$SCRIPT_DIR/create_logs.sh &
