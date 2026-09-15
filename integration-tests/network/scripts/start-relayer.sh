#!/bin/bash

set -eu
source scripts/utils.sh
source scripts/constants.sh

CHAIN_ID_A=${CHAIN_NAME_A}-test-1
CHAIN_ID_B=${CHAIN_NAME_B}-test-1
RELAYER_DEPENDENCY=${RELAYER_DEPENDENCY:-}

wait_for_node $CHAIN_NAME_A 
wait_for_node $CHAIN_NAME_B

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

    # REHEARSAL ONLY — DO NOT MERGE: with RELAYER_MANUAL=true the daemon is not started, so
    # packets can be relayed one command at a time (rly tx relay-packets / relay-acknowledgements)
    # from a kubectl exec. Toggle with `kubectl set env deployment/relayer-stride-cosmoshub RELAYER_MANUAL=true`
    if [[ "${RELAYER_MANUAL:-false}" == "true" ]]; then
        echo "RELAYER_MANUAL=true: not starting the relayer daemon"
        sleep infinity
    fi

    # REHEARSAL ONLY — DO NOT MERGE: RELAYER_DENY_CHANNELS=channel-1[,channel-2] keeps the daemon
    # off those channels (from the stride side) so they can be relayed by hand while every other
    # channel — and the light clients — stay serviced. Toggle with
    # `kubectl set env deployment/relayer-stride-cosmoshub RELAYER_DENY_CHANNELS=channel-1`
    if [[ -n "${RELAYER_DENY_CHANNELS:-}" ]]; then
        echo "Denying channels ${RELAYER_DENY_CHANNELS} on ${PATH_NAME}"
        rly paths update $PATH_NAME --filter-rule denylist --filter-channels "$RELAYER_DENY_CHANNELS"
    fi

    echo "Starting relayer..."
    # REHEARSAL ONLY: one msg per tx — batched MsgTimeouts on an ordered channel fail proof verification
    rly start $PATH_NAME --max-msgs 1
}

main