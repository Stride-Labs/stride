#!/bin/bash
set -eu

# Wait for API server to start
wait_for_api() {
    api_endpoint="$1"
    until [[ $(curl -o /dev/null -H "Cache-Control: no-cache" -s -w "%{http_code}\n" "${api_endpoint}/status") -eq 200 ]]; do
        echo "Waiting for API to start..."
        sleep 2
    done
}

# Wait for node to start
wait_for_node() {
    chain_name="$1"
    rpc_endpoint="http://${chain_name}-validator.integration.svc:26657/status"

    # Wait for the node to be caught up and confirm it's at least on the 2nd block
    until 
        response=$(curl -s --connect-timeout 5 -H "Cache-Control: no-cache" "$rpc_endpoint")
        catching_up=$(echo "$response" | jq -r '.result.sync_info.catching_up')
        latest_block=$(echo "$response" | jq -r '.result.sync_info.latest_block_height')
        [[ $catching_up == "false" && $latest_block -gt 2 ]]
    do
        echo "Waiting for $chain_name to start..."
        sleep 2
    done 
}