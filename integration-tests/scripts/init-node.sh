#!/bin/bash

set -eu 
source scripts/config.sh

# Wait for API server to start
wait_for_api $API_ENDPOINT

# Initialize the config directory and validator key if it's not the main node
init_config() {
    if [[ "$VALIDATOR_INDEX" != "1" ]]; then
        moniker=${CHAIN_NAME}${VALIDATOR_INDEX}
        $BINARY init $moniker --chain-id $CHAIN_ID --overwrite 
        $BINARY config keyring-backend test
    fi
}

# Helper function to download a file from the API
download_shared_file() {
    stored_path="$1"
    destination_path="$2"

    status_code=$(curl -s -o $destination_path -w "%{http_code}" "${API_ENDPOINT}/download/${stored_path}")
    if [[ "$status_code" != "200" ]]; then
        echo "ERROR - Failed to download $stored_path, status code ${status_code}"
        exit 1
    fi
}

# Update config.toml, app.toml, and client.toml
update_config() {
    config_toml="${CHAIN_HOME}/config/config.toml"
    client_toml="${CHAIN_HOME}/config/client.toml"
    app_toml="${CHAIN_HOME}/config/app.toml"

    echo "Updating config.toml..."
    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" $config_toml
    sed -i -E "s|127.0.0.1|0.0.0.0|g" $config_toml
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" $config_toml
    sed -i -E "s|prometheus = false|prometheus = true|g" $config_toml

    echo "Updating app.toml..."
    sed -i -E "s|minimum-gas-prices = \".*\"|minimum-gas-prices = \"0${DENOM}\"|g" $app_toml
    sed -i -E '/\[api\]/,/^enable = .*$/ s/^enable = .*$/enable = true/' $app_toml
    sed -i -E 's|unsafe-cors = .*|unsafe-cors = true|g' $app_toml
    sed -i -E 's|localhost|0.0.0.0|g' $app_toml

    echo "Updating client.toml"
    sed -i -E "s|chain-id = \"\"|chain-id = \"${CHAIN_ID}\"|g" $client_toml
    sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $client_toml
    sed -i -E "s|node = \".*\"|node = \"tcp://localhost:${RPC_PORT}\"|g" $client_toml

    echo "Retrieving private keys and genesis.json..."
    download_shared_file ${VALIDATOR_KEYS_DIR}/${CHAIN_NAME}/val${VALIDATOR_INDEX}.json ${CHAIN_HOME}/config/priv_validator_key.json 
    download_shared_file ${NODE_KEYS_DIR}/${CHAIN_NAME}/val${VALIDATOR_INDEX}.json  ${CHAIN_HOME}/config/node_key.json 
    download_shared_file ${GENESIS_DIR}/${CHAIN_NAME}/genesis.json ${CHAIN_HOME}/config/genesis.json 
}

# Update the persistent peers conditionally based on which node it is
add_peers() {
    echo "Setting peers..."
    if [[ "$VALIDATOR_INDEX" == "1" ]]; then
        # For the main node, wipe out the persistent peers that are incorrectly generated 
        sed -i -E "s|^persistent_peers = .*|persistent_peers = \"\"|g" $config_toml
    else
        # For the other nodes, add the main node as the persistent peer
        download_shared_file ${NODE_IDS_DIR}/${CHAIN_NAME}/val1.txt main_node_id.txt
        main_node_id=$(cat main_node_id.txt)
        main_pod_id=${CHAIN_NAME}-validator-0
        service=${CHAIN_NAME}-validator
        persistent_peer=${main_node_id}@${main_pod_id}.${service}.${NAMESPACE}.svc.cluster.local:${PEER_PORT}
        sed -i -E "s|^persistent_peers = .*|persistent_peers = \"${persistent_peer}\"|g" $config_toml
    fi
}

main() {
    echo "Initializing node..."
    init_config
    update_config
    add_peers
    echo "Done"
}

main
