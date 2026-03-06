#!/bin/bash
# Validates update-msgs JSON files against on-chain validator sets.
# Usage: ./validate.sh [chain-id]
# If no chain-id is provided, validates all chains.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
UPDATE_DIR="$SCRIPT_DIR/update-msgs"
ADD_DIR="$SCRIPT_DIR/add-msgs"

validate_chain() {
    local chain_id="$1"
    local json_file="$UPDATE_DIR/$chain_id.json"
    local add_file="$ADD_DIR/$chain_id.json"

    if [ ! -f "$json_file" ]; then
        echo "$chain_id: no update msg file found"
        return 1
    fi

    # Get on-chain validator addresses
    onchain_addrs=$(strided q stakeibc show-host-zone "$chain_id" --output json \
        | jq -r '.host_zone.validators[].address' | sort)

    # Get addresses that will be added (these will be on-chain after the add tx)
    add_addrs=""
    if [ -f "$add_file" ]; then
        add_addrs=$(jq -r '.validators[].address' "$add_file" | sort)
    fi

    # Effective on-chain = current on-chain + pending adds
    effective_addrs=$(printf '%s\n%s' "$onchain_addrs" "$add_addrs" | grep -v '^$' | sort -u)

    # Get addresses from the update JSON
    json_addrs=$(jq -r '.validator_weights[].address' "$json_file" | sort)

    # Find addresses in JSON but not in effective on-chain set
    not_onchain=$(comm -23 <(echo "$json_addrs") <(echo "$effective_addrs"))
    # Find addresses in effective on-chain set but not in JSON
    not_in_json=$(comm -13 <(echo "$json_addrs") <(echo "$effective_addrs"))

    local errors=0

    if [ -n "$not_onchain" ]; then
        count=$(echo "$not_onchain" | wc -l | tr -d ' ')
        echo "$chain_id: ERROR - $count validators in JSON but NOT on-chain (even after adds):"
        echo "$not_onchain" | sed 's/^/  /'
        errors=1
    fi

    if [ -n "$not_in_json" ]; then
        count=$(echo "$not_in_json" | wc -l | tr -d ' ')
        echo "$chain_id: ERROR - $count validators on-chain but NOT in JSON:"
        echo "$not_in_json" | sed 's/^/  /'
        errors=1
    fi

    if [ "$errors" -eq 0 ]; then
        json_count=$(echo "$json_addrs" | wc -l | tr -d ' ')
        effective_count=$(echo "$effective_addrs" | wc -l | tr -d ' ')
        echo "$chain_id: OK ($json_count in JSON, $effective_count on-chain after adds)"
    fi

    return $errors
}

if [ -n "$1" ]; then
    validate_chain "$1"
else
    has_errors=0
    for json_file in "$UPDATE_DIR"/*.json; do
        chain_id=$(basename "$json_file" .json)
        validate_chain "$chain_id" || has_errors=1
    done
    exit $has_errors
fi
