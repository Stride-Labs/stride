#!/bin/bash
set -eu

# Wait for API server to start
wait_for_api() {
    api_endpoint="$1"
    until [[ $(curl -o /dev/null -s -w "%{http_code}\n" "${api_endpoint}/status") -eq 200 ]]; do
        echo "Waiting for API to start..."
        sleep 2
    done
}

# Wait for node to start
wait_for_node() {
    chain_name="$1"
    rpc_endpoint="http://${chain_name}-validator.integration.svc:26657/status"
    until [[ $(curl -s "$rpc_endpoint" | jq '.result.sync_info.catching_up') == "false" ]]; do
        echo "Waiting for $chain_name to start..."
        sleep 2
    done 
}