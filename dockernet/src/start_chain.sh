#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

for chain_id in STRIDE ${HOST_CHAINS[@]} ${ACCESSORY_CHAINS[@]:-}; do
    num_nodes=$(GET_VAR_VALUE ${chain_id}_NUM_NODES)
    node_prefix=$(GET_VAR_VALUE ${chain_id}_NODE_PREFIX)

    log_file=$DOCKERNET_HOME/logs/${node_prefix}.log

    echo "Starting $chain_id chain"
    nodes_names=$(i=1; while [ $i -le $num_nodes ]; do printf "%s " ${node_prefix}${i}; i=$(($i + 1)); done;)
    $DOCKER_COMPOSE up -d $nodes_names

    SAVE_DOCKER_LOGS ${node_prefix}1 $log_file
done

for chain_id in STRIDE ${HOST_CHAINS[@]} ${ACCESSORY_CHAINS[@]:-}; do
    printf "Waiting for $chain_id to start... "

    node_prefix=$(GET_VAR_VALUE ${chain_id}_NODE_PREFIX)
    log_file=$DOCKERNET_HOME/logs/${node_prefix}.log

    ( tail -f -n0 $log_file & ) | grep -q "finalizing commit of block"
    echo "Done"
done

sleep 5