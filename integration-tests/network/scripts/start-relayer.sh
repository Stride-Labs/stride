#!/bin/bash

set -eu
source scripts/utils.sh
source scripts/constants.sh

wait_for_node $CHAIN_NAME_A 
wait_for_node $CHAIN_NAME_B

restore_keys() {
    mnemonic_a=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_A_MNEMONIC_INDEX" ${KEYS_FILE})
    mnemonic_b=$(jq -r '.relayers[$index].mnemonic' --argjson index "$CHAIN_B_MNEMONIC_INDEX" ${KEYS_FILE})

    rly keys restore $CHAIN_NAME_A $CHAIN_NAME_A "$mnemonic_a"
    rly keys restore $CHAIN_NAME_B $CHAIN_NAME_B "$mnemonic_b"
}

main() {
    relayer_config_file=${HOME}/.relayer/config/config.yaml

    # The config is mounted from a configmap which is read-only by default
    # In order to make it writeable, we need to copy it to a new location
    # This also acts as an indicator as to whether the relayer's already been started
    if [ ! -e "${relayer_config_file}" ]; then
        mkdir -p $(dirname $relayer_config_file)
        cp configs/relayer.yaml ${relayer_config_file}

        restore_keys
        rly tx link $PATH_NAME
        rly start $PATH_NAME
    else 
        restore_keys
        rly start $PATH_NAME
    fi
}

main