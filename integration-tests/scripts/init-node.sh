#!/bin/bash

set -eu 

BINARY=strided
CHAIN_NAME=stride
CHAIN_HOME=.stride
DENOM=ustrd
BLOCK_TIME=1s
CHAIN_ID=${CHAIN_NAME}-test-1
STAKE_TOKENS=1000000
MICRO_DENOM_UNITS=000000

PEER_PORT=26656
RPC_PORT=26657
POD_INDEX=${POD_NAME##*-}
VALIDATOR_INDEX=$((POD_INDEX+1))
VALIDATOR_NAME=val${VALIDATOR_INDEX}
SHARED_DIR=shared

config_toml="${CHAIN_HOME}/config/config.toml"
client_toml="${CHAIN_HOME}/config/client.toml"
app_toml="${CHAIN_HOME}/config/app.toml"

# Initialize the config directory and validator key if it's not the main node
init_config() {
    if [[ "$VALIDATOR_INDEX" != "1" ]]; then
        moniker=${CHAIN_NAME}${VALIDATOR_INDEX}
        $BINARY init $moniker --chain-id $CHAIN_ID --overwrite 
        $BINARY config keyring-backend test

        validator_config=$(jq -r '.validators[$index]' --argjson index "$POD_INDEX" config/keys.json)
        mnemonic=$(echo $validator_config | jq -r '.mnemonic')
        name=$(echo $validator_config | jq -r '.name')

        # echo "$mnemonic" | $BINARY keys add $name --recover 
    fi
}

# Update config.toml, app.toml, and client.toml
update_config() {
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
}

# Extract private keys and genesis
download_shared() {
    echo "Retrieving private keys and genesis.json..."
    ls -la ${SHARED_DIR}
    ls -la ${SHARED_DIR}/validator-keys
    cp ${SHARED_DIR}/validator-keys/val${VALIDATOR_INDEX}.json ${CHAIN_HOME}/config/priv_validator_key.json
    cp ${SHARED_DIR}/node-keys/val${VALIDATOR_INDEX}.json ${CHAIN_HOME}/config/node_key.json
    cp ${SHARED_DIR}/genesis.json ${CHAIN_HOME}/config/genesis.json
}

# Update the persistent peers conditionally based on which node it is
add_peers() {
    echo "Setting peers..."
    if [[ "$VALIDATOR_INDEX" == "1" ]]; then
        # For the main node, wipe out the persistent peers that are incorrectly generated 
        sed -i -E "s|^persistent_peers = .*|persistent_peers = \"\"|g" $config_toml
    else
        # For the other nodes, add the main node as the persistent peer
        main_node_id=$(cat ${SHARED_DIR}/node-ids/val1.txt)
        main_pod_id=${CHAIN_NAME}-validator-0
        service=${CHAIN_NAME}-validator
        persistent_peer=${main_node_id}@${main_pod_id}.${service}.${NAMESPACE}.svc.cluster.local:${PEER_PORT}
        sed -i -E "s|^persistent_peers = .*|persistent_peers = \"${persistent_peer}\"|g" $config_toml
    fi
}

# Create the governor
create_governor() {
    echo "Creating governor..."
    pub_key=$($BINARY tendermint show-validator)
    $BINARY tx staking create-validator --amount ${STAKE_TOKENS}${MICRO_DENOM_UNITS}${DENOM} --from ${VALIDATOR_NAME} \
        --pubkey=$pub_key --commission-rate="0.10" --commission-max-rate="0.20" \
        --commission-max-change-rate="0.01" --min-self-delegation="1" -y
}

echo "Initializing node..."
init_config
update_config
download_shared
add_peers
# create_governor
echo "Done"