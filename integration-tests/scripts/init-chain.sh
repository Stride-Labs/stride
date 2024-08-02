#!/bin/bash

set -eu
source scripts/config.sh

LOCAL_MODE=${1:-false}

# If this is being run locally, don't overwrite the main chain folder
if [[ "$LOCAL_MODE" == "true" ]]; then
    CHAIN_HOME=state
    rm -rf state
    BINARY="$BINARY --home $CHAIN_HOME"
    API_ENDPOINT="http://localhost:8000"
fi

# Wait for API server to start
wait_for_api $API_ENDPOINT

# check if the binary has genesis subcommand or not, if not, set CHAIN_GENESIS_CMD to empty
genesis_json=${CHAIN_HOME}/config/genesis.json
chain_genesis_command=$($BINARY 2>&1 | grep -q "genesis-related subcommands" && echo "genesis" || echo "")

# Helper to update a json attribute in-place
jq_inplace() {
    jq_filter="$1"
    file="$2"
    
    jq "$jq_filter" "$file" > /tmp/$(basename $file) && mv /tmp/$(basename $file) ${file}
}

# Initializes the chain config
init_config() {
    moniker=${CHAIN_NAME}1
    $BINARY init $moniker --chain-id $CHAIN_ID --overwrite 
    $BINARY config keyring-backend test
}

# Helper to upload shared files to the API
upload_shared_file() {
    file_path="$1"
    saved_path="${2:-}"
    file_name=$(basename $file_path)

    curl -s -X 'POST' "${API_ENDPOINT}/upload/${saved_path}" \
        -H 'accept: application/json' \
        -H 'Content-Type: multipart/form-data' \
        -F "file=@$file_path" && echo 
}

# Adds each validator to the genesis file, and also saves down the public keys 
# which are needed for ICS
# Each validators public private key and node ID are saved in the API
add_validators() {
    echo "Adding validators..."

    validator_public_keys=""
    for (( i=1; i <= $NUM_VALIDATORS; i++ )); do
        # Extract the validator name and mnemonic from keys.json
        validator_config=$(jq -r '.validators[$index]' --argjson index "$((i-1))" ${KEYS_FILE})
        name=$(echo $validator_config | jq -r '.name')
        mnemonic=$(echo $validator_config | jq -r '.mnemonic')

        # Add the key to the main keyring the the validator's sub-keyring
        echo "$mnemonic" | $BINARY keys add $name --recover
        address=$($BINARY keys show $name -a)

        # Use a separate directory for the non-main nodes so we can generate unique validator keys
        if [[ "$i" == "1" ]]; then 
            validator_home=${CHAIN_HOME}
        else 
            validator_home=/tmp/${CHAIN_NAME}-${name} && rm -rf $validator_home
            $BINARY init $name --chain-id $CHAIN_ID --overwrite --home ${validator_home} &> /dev/null
        fi

        # Add the genesis account 
        genesis_balance=${VALIDATOR_BALANCE}${DENOM}
        $BINARY $chain_genesis_command add-genesis-account $address $genesis_balance

        # Save the node IDs and keys to the API
        $BINARY tendermint show-node-id --home ${validator_home} > node_id.txt
        upload_shared_file node_id.txt ${NODE_IDS_DIR}/${name}.txt 
        upload_shared_file ${validator_home}/config/priv_validator_key.json ${VALIDATOR_KEYS_DIR}/${name}.json
        upload_shared_file ${validator_home}/config/node_key.json ${NODE_KEYS_DIR}/${name}.json

        # Save the comma separted public keys for the ICS genesis update
        validator_public_keys+="$(jq -r '.pub_key.value' ${validator_home}/config/priv_validator_key.json),"
    done

    # For non-stride nodes, generate and collect the validator gentx (for the main node only)
    # The other validators will be created after startup
    if [[ "$CHAIN_NAME" != "stride" ]]; then 
        $BINARY $chain_genesis_command gentx val1 ${VALIDATOR_STAKE}${DENOM} --chain-id $CHAIN_ID 
        $BINARY $chain_genesis_command collect-gentxs
    fi
}

# Updates the genesis config with defaults
update_default_genesis() {
    echo "Updating genesis.json with defaults..."

    sed -i -E "s|\"stake\"|\"${DENOM}\"|g" $genesis_json 
    sed -i -E "s|\"aphoton\"|\"${DENOM}\"|g" $genesis_json # ethermint default

    jq_inplace '.app_state.staking.params.unbonding_time |= "'$UNBONDING_TIME'"' $genesis_json
    jq_inplace '.app_state.gov.params.max_deposit_period |= "'$DEPOSIT_PERIOD'"' $genesis_json 
    jq_inplace '.app_state.gov.params.voting_period |= "'$VOTING_PERIOD'"' $genesis_json 

    if jq 'has(.app_state.gov.params.expedited_voting_period)' $genesis_json > /dev/null 2>&1; then
        jq_inplace '.app_state.gov.params.expedited_voting_period |= "'$EXPEDITED_VOTING_PERIOD'"' $genesis_json 
    fi
}

# Genesis updates specific to stride
update_stride_genesis() {
    echo "Updating genesis.json with stride configuration..."

    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration |= "'$STRIDE_DAY_EPOCH_DURATION'"' $genesis_json 
    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration |= "'$STRIDE_EPOCH_EPOCH_DURATION'"' $genesis_json 

    $BINARY add-consumer-section --validator-public-keys $validator_public_keys
}

# Genesis updates specific to non-stride chains
update_host_genesis() {
    echo "Updating genesis.json with host configuration..."
}

# Saves the genesis file in the API
save_genesis() {
    echo "Saving genesis.json..."

    upload_shared_file $genesis_json genesis.json
}

main() {
    echo "Initializing chain..."
    init_config
    add_validators
    update_default_genesis
    if [[ "$CHAIN_NAME" == "stride" ]]; then 
        update_stride_genesis
    else 
        update_host_genesis
    fi
    save_genesis
    echo "Done"
}

main