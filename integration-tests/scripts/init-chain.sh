#!/bin/bash

set -e

LOCAL_MODE=${1:-false}

CHAIN_NAME=stride
BINARY=strided
DENOM=ustrd
CHAIN_ID=${CHAIN_NAME}-test-1
VALIDATOR_BALANCE=10000000
MICRO_DENOM_UNITS=000000
CHAIN_HOME=.stride
SHARED_DIR=shared
NUM_VALIDATORS=3

STRIDE_DAY_EPOCH_DURATION="140s"
STRIDE_EPOCH_EPOCH_DURATION="35s"
MAX_DEPOSIT_PERIOD="30s"
VOTING_PERIOD="30s"
UNBONDING_TIME="240s"

# If this is being run locally, don't overwrite the main chain folder
if [[ "$LOCAL_MODE" == "true" ]]; then
    CHAIN_HOME=state
    rm -rf state
    BINARY="$BINARY --home $CHAIN_HOME"
fi

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

# Adds each validator to the genesis file, and also saves down the public keys 
# which are needed for ICS
# Each validators public private key and node ID are saved in the shared directory
add_validators() {
    echo "Adding validators..."

    # check if the binary has genesis subcommand or not, if not, set CHAIN_GENESIS_CMD to empty
    chain_genesis_command=$($BINARY 2>&1 | grep -q "genesis-related subcommands" && echo "genesis" || echo "")

    validator_public_keys=""
    for (( i=1; i <= $NUM_VALIDATORS; i++ )); do
        # Extract the validator name and mnemonic from keys.json
        validator_config=$(jq -r '.validators[$index]' --argjson index "$((i-1))" config/keys.json)
        name=$(echo $validator_config | jq -r '.name')
        mnemonic=$(echo $validator_config | jq -r '.mnemonic')

        # Add the key to the keyring
        echo "$mnemonic" | $BINARY keys add $name --recover 
        address=$($BINARY keys show $name -a)

        # Add the genesis account
        genesis_balance=${VALIDATOR_BALANCE}${MICRO_DENOM_UNITS}${DENOM}
        $BINARY $chain_genesis_command add-genesis-account $address $genesis_balance

        # Save the node-id and validator private keys to the shared directory
        validator_home=/tmp/${CHAIN_NAME}-${name}
        $BINARY init $CHAIN_NAME-$name --chain-id $CHAIN_ID --overwrite --home ${validator_home} &> /dev/null
        node_id=$($BINARY tendermint show-node-id --home ${validator_home})

        mkdir -p ${SHARED_DIR}/validator-keys
        mkdir -p ${SHARED_DIR}/node-keys
        mkdir -p ${SHARED_DIR}/node-ids
    
        mv ${validator_home}/config/priv_validator_key.json ${SHARED_DIR}/validator-keys/${name}.json
        mv ${validator_home}/config/node_key.json ${SHARED_DIR}/node-keys/${name}.json
        echo $node_id > ${SHARED_DIR}/node-ids/${name}.txt

        # Save the comma separted public keys for the ICS genesis update
        validator_public_keys+="$(jq -r '.pub_key.value' ${SHARED_DIR}/validator-keys/${name}.json),"
    done
}

# Updates the genesis config with defaults
update_genesis() {
    echo "Updating genesis.json"

    genesis_json=${CHAIN_HOME}/config/genesis.json

    sed -i -E "s|\"stake\"|\"${DENOM}\"|g" $genesis_json 
    sed -i -E "s|\"aphoton\"|\"${DENOM}\"|g" $genesis_json # ethermint default

    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration |= "'$STRIDE_DAY_EPOCH_DURATION'"' $genesis_json 
    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration |= "'$STRIDE_EPOCH_EPOCH_DURATION'"' $genesis_json 
    jq_inplace '.app_state.gov.params.max_deposit_period |= "'$MAX_DEPOSIT_PERIOD'"' $genesis_json 
    jq_inplace '.app_state.gov.params.voting_period |= "'$VOTING_PERIOD'"' $genesis_json 

    $BINARY add-consumer-section --validator-public-keys $validator_public_keys

    cp $genesis_json ${SHARED_DIR}/genesis.json
}

echo "Initializing chain..."
init_config
add_validators
update_genesis
echo "Done"
