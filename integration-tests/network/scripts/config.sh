#!/bin/bash

# Override the hostname only when in local mode
if [[ "$HOSTNAME" != *"validator"* ]]; then
    HOSTNAME=validator-0
fi

SCRIPTS_DIR=scripts
CONFIG_DIR=configs

VALIDATOR_KEYS_DIR=validator-keys
NODE_KEYS_DIR=node-keys
NODE_IDS_DIR=node-ids
GENESIS_DIR=genesis
KEYS_FILE=${CONFIG_DIR}/keys.json

POD_INDEX=${HOSTNAME##*-}
VALIDATOR_INDEX=$((POD_INDEX+1))
VALIDATOR_NAME=val${VALIDATOR_INDEX}

API_ENDPOINT=http://api.integration.svc:8000

PEER_PORT=26656
RPC_PORT=26657

# Redefined to confirm they're set
CHAIN_NAME=${CHAIN_NAME}
CHAIN_HOME=${CHAIN_HOME}
BINARY=${BINARY}
DENOM=${DENOM}
DENOM_DECIMALS=${DENOM_DECIMALS}
NUM_VALIDATORS=${NUM_VALIDATORS}

MICRO_DENOM_ZERO_PAD=$(printf "%${DENOM_DECIMALS}s" | tr ' ' "0")
CHAIN_ID=${CHAIN_NAME}-test-1
BLOCK_TIME=1s
VALIDATOR_BALANCE=10000000${MICRO_DENOM_ZERO_PAD}
VALIDATOR_STAKE=1000000${MICRO_DENOM_ZERO_PAD}

DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
EXPEDITED_VOTING_PERIOD="29s"
UNBONDING_TIME="240s"

STRIDE_DAY_EPOCH_DURATION="140s"
STRIDE_EPOCH_EPOCH_DURATION="35s"

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
    until [[ $(curl -o /dev/null -s $rpc_endpoint | jq '.result.sync_info.catching_up') == "false" ]]; do
        echo "Waiting for $chain_name to start..."
        sleep 2
    done 
}