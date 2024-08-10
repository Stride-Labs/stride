#!/bin/bash

set -eu
source scripts/utils.sh
source scripts/constants.sh

CHAIN_ID_A=${CHAIN_NAME_A}-test-1
CHAIN_ID_B=${CHAIN_NAME_B}-test-1

wait_for_node $CHAIN_NAME_A 
wait_for_node $CHAIN_NAME_B

restore_keys() {
    mnemonic_a=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_A_MNEMONIC_INDEX" ${KEYS_FILE})
    mnemonic_b=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_B_MNEMONIC_INDEX" ${KEYS_FILE})

    echo "$mnemonic_a" > /tmp/mnemonics_a.txt
    echo "$mnemonic_b" > /tmp/mnemonics_b.txt

    hermes keys add --chain $CHAIN_ID_A --mnemonic-file /tmp/mnemonics_a.txt 
    hermes keys add --chain $CHAIN_ID_B --mnemonic-file /tmp/mnemonics_b.txt 

    rm -f /tmp/mnemonics_a.txt /tmp/mnemonics_b.txt 
}

create_path() {
    # Only create a path if one does not already exist
    if ! hermes query channels --chain $CHAIN_ID_A --show-counterparty | grep -q $CHAIN_ID_A; then
        echo "Creating path..."
        hermes create channel --a-chain $CHAIN_ID_A --b-chain $CHAIN_ID_B \
            --a-port transfer --b-port transfer --new-client-connection --yes
    fi
}

main() {
    # The config is mounted from a configmap which is read-only by default
    # In order to make it writeable, we need to copy it to a new location
    hermes_config_file=${HOME}/.hermes/config.toml
    mkdir -p $(dirname $hermes_config_file)
    cp configs/hermes.toml ${hermes_config_file}

    restore_keys
    create_path

    echo "Starting hermes..."
    hermes start
}

main