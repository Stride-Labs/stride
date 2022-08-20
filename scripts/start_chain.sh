#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

# Pass the CHAIN_ID's of the chains to start
CHAINS="$@"

for chain_id in ${CHAINS[@]}; do
    __num_nodes_var=${chain_id}_NUM_NODES
    __node_prefix_var=${chain_id}_NODE_PREFIX

    num_nodes=${!__num_nodes_var}
    node_prefix=${!__node_prefix_var}

    log_file=$SCRIPT_DIR/logs/${node_prefix}.log

    echo "Starting $chain_id chain"
    nodes_names=$(i=1; while [ $i -le $num_nodes ]; do printf "%s " ${node_prefix}${i}; i=$(($i + 1)); done;)
    docker-compose up -d $nodes_names

    docker-compose logs -f ${node_prefix}1 | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $log_file 2>&1 &
done

for chain_id in ${CHAINS[@]}; do
    printf "Waiting for $chain_id to start..."

    __node_prefix_var=${chain_id}_NODE_PREFIX
    node_prefix=${!__node_prefix_var}
    
    log_file=$SCRIPT_DIR/logs/${node_prefix}.log

    ( tail -f -n0 $log_file & ) | grep -q "finalizing commit of block"
    echo "Done"
done

sleep 5