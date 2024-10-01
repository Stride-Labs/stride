#!/bin/bash

set -e

relayer_type="$1"

chain_id="${CHAIN_NAME_A}-test-1"
relayer_config_file=${HOME}/.relayer/config/config.yaml
hermes_config_file=${HOME}/.hermes/config.toml

if [[ "$relayer_type" == "relayer" ]]; then
    if [[ ! -f $relayer_config_file ]]; then
        echo "Config not initialized yet"
        exit 1
    fi
    if ! rly q channels "$CHAIN_NAME_A" 2>/dev/null | grep -q STATE_OPEN; then
        echo "Source channel not open yet"
        exit 1
    fi
    if ! rly q channels "$CHAIN_NAME_B" 2>/dev/null | grep -q STATE_OPEN; then
        echo "Destination channel not open yet"
        exit 1
    fi
    exit 0

elif [[ "$relayer_type" == "hermes" ]]; then
    if [[ ! -f $hermes_config_file ]]; then
        echo "Config not initialized yet"
        exit 1
    fi
    open_channels=$(hermes query channels --chain "$chain_id" --show-counterparty 2>/dev/null | grep -o "channel-" | wc -l)
    if [[ "$open_channels" == "0" ]]; then 
        echo "Source channel not open yet"
        exit 1
    fi
    if [[ "$open_channels" == "1" ]]; then 
        echo "Destination channel not open yet"
        exit 1
    fi
    exit 0

else
    echo "ERROR: Unsupported relayer type $relayer_type"
    exit 1
fi
