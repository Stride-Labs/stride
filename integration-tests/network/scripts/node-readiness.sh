#!/bin/bash

set -e
source scripts/config.sh

# If chain hasn't been initialized yet, exit immediately
if [ ! -d $CHAIN_HOME/config ]; then
    echo "READINESS CHECK FAILED - Chain has not been initialized yet."
    exit 1
fi

# Check that the node is running
if ! $($BINARY status &> /dev/null); then
    echo "READINESS CHECK FAILED - Chain is down"
    exit 1
fi

# It's not possible for one node to start up by itself (without peers), 
# so if we identify that the node is on block 0, we'll mark it as ready
# so the other nodes can start connecting
if [[ "$($BINARY status 2>/dev/null | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end')" == "0" ]]; then
    exit 0
fi

# Then check if the node is synced according to it's status query
CATCHING_UP=$($BINARY status 2>&1 | jq "if .SyncInfo then .SyncInfo.catching_up else .sync_info.catching_up end")
if [[ "$CATCHING_UP" != "false" ]]; then
    echo "READINESS CHECK FAILED - Node is still syncing"
    exit 1
fi