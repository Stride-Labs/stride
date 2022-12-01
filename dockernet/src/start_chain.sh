#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

for chain_id in STRIDE ${HOST_CHAINS[@]}; do
    num_nodes=$(GET_VAR_VALUE ${chain_id}_NUM_NODES)
    node_prefix=$(GET_VAR_VALUE ${chain_id}_NODE_PREFIX)

    log_file=$SCRIPT_DIR/logs/${node_prefix}.log

    echo "Starting $chain_id chain"
    nodes_names=$(i=1; while [ $i -le $num_nodes ]; do printf "%s " ${node_prefix}${i}; i=$(($i + 1)); done;)
    $DOCKER_COMPOSE up -d $nodes_names

    $DOCKER_COMPOSE logs -f ${node_prefix}1 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $log_file 2>&1 &
done

for chain_id in STRIDE ${HOST_CHAINS[@]}; do
    printf "Waiting for $chain_id to start..."

    node_prefix=$(GET_VAR_VALUE ${chain_id}_NODE_PREFIX)
    log_file=$SCRIPT_DIR/logs/${node_prefix}.log

    ( tail -f -n0 $log_file & ) | grep -q "finalizing commit of block"
    echo "Done"
done

sleep 5