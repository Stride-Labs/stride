#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
docker-compose down
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp
mkdir -p $SCRIPT_DIR/logs

# Initialize the state for stride/gaia and relayers
sh ${SCRIPT_DIR}/init_chain.sh STRIDE
sh ${SCRIPT_DIR}/init_chain.sh GAIA
sh ${SCRIPT_DIR}/init_chain.sh JUNO
sh ${SCRIPT_DIR}/init_chain.sh OSMO

HOST_CHAINS=(GAIA JUNO OSMO)
sh ${SCRIPT_DIR}/start_chain.sh STRIDE ${HOST_CHAINS[@]}
sh ${SCRIPT_DIR}/start_relayers.sh ${HOST_CHAINS[@]} 

exit 0

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