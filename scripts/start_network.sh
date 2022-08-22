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

HOST_CHAINS=(GAIA JUNO OSMO)
sh ${SCRIPT_DIR}/start_chain.sh STRIDE ${HOST_CHAINS[@]}
sh ${SCRIPT_DIR}/init_relayers.sh STRIDE ${HOST_CHAINS[@]}
sh ${SCRIPT_DIR}/create_channels.sh ${HOST_CHAINS[@]}

# printf "Creating clients, connections, and transfer channel"
# $RELAYER_EXEC transact link stride-gaia
# echo "DONE"

echo "Starting relayers"
docker-compose up -d hermes icq

docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 &
docker-compose logs -f icq | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $ICQ_LOGS 2>&1 &

( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "Hermes has started"

# Register all host zones in parallel
pids=()
for i in ${!HOST_CHAINS[@]}; do
    if [[ "$i" != "0" ]]; then sleep 20; fi
    bash $SCRIPT_DIR/register_host.sh ${HOST_CHAINS[$i]} connection-${i} channel-${i} &
    pids[${i}]=$!
done
for i in ${!pids[@]}; do
    wait ${pids[$i]}
    echo "${HOST_CHAINS[$i]} - Done"
done

$SCRIPT_DIR/create_logs.sh &
