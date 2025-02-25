#!/bin/bash

set -eu
source scripts/utils.sh
source scripts/constants.sh

CHAIN_ID_A=${CHAIN_NAME_A}-test-1
# CHAIN_ID_B=${CHAIN_NAME_B}-test-1
CHAIN_ID_B=osmosis-1

RELAYER_DEPENDENCY=${RELAYER_DEPENDENCY:-}

wait_for_node $CHAIN_NAME_A 
# wait_for_node $CHAIN_NAME_B

relayer_config_file=${HOME}/.relayer/config/config.yaml

restore_keys() {
    mnemonic_a=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_A_MNEMONIC_INDEX" ${KEYS_FILE})
    mnemonic_b=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_B_MNEMONIC_INDEX" ${KEYS_FILE})

    rly keys restore $CHAIN_NAME_A $CHAIN_NAME_A "$mnemonic_a"
    rly keys restore $CHAIN_NAME_B $CHAIN_NAME_B "$mnemonic_b"
}

wait_for_turn() {
    if [[ -z "$RELAYER_DEPENDENCY" ]]; then 
        echo "First relayer in sequence"
        return
    fi
    until nslookup ${RELAYER_DEPENDENCY}.${NAMESPACE}.svc.cluster.local; do 
        echo Waiting for $RELAYER_DEPENDENCY to start...
        sleep 2 
    done
}

create_path() {
    # If there aren't any channels between the two chains yet, create a new path
    if ! rly q channels $CHAIN_NAME_A $CHAIN_NAME_B 2>/dev/null | grep -q STATE_OPEN; then
        echo "Creating path..."
        rly tx link $PATH_NAME
    else
        # Otherwise, add the existing connection to the config
        client_id_a=$(rly q clients $CHAIN_NAME_A | grep $CHAIN_ID_B | jq -r '.client_id')
        client_id_b=$(rly q clients $CHAIN_NAME_B | grep $CHAIN_ID_A | jq -r '.client_id')

        connection_id_a=$(rly q client-connections $CHAIN_NAME_A $client_id_a | jq -r '.connections[0].id')
        connection_id_b=$(rly q client-connections $CHAIN_NAME_B $client_id_b | jq -r '.connections[0].id')

        echo "Path already found, updating config..."
        yq eval -i "
            .paths.\"${PATH_NAME}\".src.client-id = \"$client_id_a\" |
            .paths.\"${PATH_NAME}\".src.connection = \"$connection_id_a\" |
            .paths.\"${PATH_NAME}\".dst.client-id = \"$client_id_b\" |
            .paths.\"${PATH_NAME}\".dst.connection = \"$connection_id_b\"
            " "$relayer_config_file"
    fi
}

main() {
    # The config is mounted from a configmap which is read-only by default
    # In order to make it writeable, we need to copy it to a new location
    mkdir -p $(dirname $relayer_config_file)
    cp configs/relayer.yaml ${relayer_config_file}

    restore_keys
    wait_for_turn
    create_path

    echo "Starting relayer..."
    rly start $PATH_NAME
}

main