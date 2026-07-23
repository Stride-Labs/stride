#!/bin/bash

set -eu
source scripts/config.sh
source scripts/utils.sh

# Wait for API server to start
wait_for_api $API_ENDPOINT

# check if the binary has genesis subcommand or not, if not, set CHAIN_GENESIS_CMD to empty
genesis_json=${CHAIN_HOME}/config/genesis.json
chain_genesis_command=$($BINARY --help 2>&1  | grep -q "genesis-related subcommands" && echo "genesis" || echo "")
client_config_command=$($BINARY config --help 2>&1  | grep -q "Set an application config" && echo "config set client" || echo "config")

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
    $BINARY $client_config_command chain-id $CHAIN_ID
    $BINARY $client_config_command keyring-backend test
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

# Helper function to add an account to the keyring and genesis
# The key arg is a JSON string with "name" and "mnemonic" fields
add_genesis_account() {
    key="$1"
    balance="$2"

    echo $key
    name=$(echo $key | jq -r '.name')
    mnemonic=$(echo $key | jq -r '.mnemonic')

    echo "$mnemonic" | $BINARY keys add $name --recover
    address=$($BINARY keys show $name -a)

    $BINARY $chain_genesis_command add-genesis-account $address $balance
}

# Adds each validator to the genesis file, generates per-validator priv_val keys
# in a separate home directory, and accumulates a POA-genesis-shaped JSON array
# of validator entries which is later injected into app_state.poa.validators.
add_validators() {
    echo "Adding validators..."

    poa_validators="[]"
    for (( i=1; i <= $NUM_VALIDATORS; i++ )); do
        # Add the validator account to the keyring and genesis
        validator_key=$(jq -r '.validators[$index]' --argjson index "$((i-1))" ${KEYS_FILE})
        name=$(echo $validator_key | jq -r '.name')

        add_genesis_account "$validator_key" "${VALIDATOR_BALANCE}${DENOM}"

        # Use a separate directory for the non-main nodes so we can generate unique validator keys
        if [[ "$i" == "1" ]]; then
            validator_home=${CHAIN_HOME}
        else
            validator_home=/tmp/${CHAIN_NAME}-${name} && rm -rf $validator_home
            $BINARY init $name --chain-id $CHAIN_ID --overwrite --home ${validator_home} &> /dev/null
        fi

        # Save the node IDs and keys to the API
        $BINARY tendermint show-node-id --home ${validator_home} > node_id.txt
        upload_shared_file node_id.txt ${NODE_IDS_DIR}/${CHAIN_NAME}/${name}.txt
        upload_shared_file ${validator_home}/config/priv_validator_key.json ${VALIDATOR_KEYS_DIR}/${CHAIN_NAME}/${name}.json
        upload_shared_file ${validator_home}/config/node_key.json ${NODE_KEYS_DIR}/${CHAIN_NAME}/${name}.json

        # Build a POA validator entry: consensus pubkey from this node's home,
        # operator address from the validator's keyring entry on the main home,
        # moniker = val name, power = "1" (equal-weight POA per the design spec).
        if [[ "$CHAIN_NAME" == "stride" ]]; then
            val_pubkey_json=$($BINARY tendermint show-validator --home ${validator_home} 2>/dev/null)
            val_op_addr=$($BINARY keys show $name -a)
            poa_validators=$(echo "$poa_validators" | jq \
                --argjson pk "$val_pubkey_json" \
                --arg op "$val_op_addr" \
                --arg moniker "$name" \
                '. + [{pub_key: $pk, power: "1", metadata: {operator_address: $op, moniker: $moniker}}]')
        fi
    done

    # Stash the POA validators array as a global so update_stride_genesis can use it.
    POA_VALIDATORS_JSON="$poa_validators"

    # For non-stride nodes, generate the validator gentx (for the main node only).
    # The other validators will be created after startup via tx staking create-validator.
    if [[ "$CHAIN_NAME" != "stride" ]]; then
        $BINARY $chain_genesis_command gentx val1 ${VALIDATOR_STAKE}${DENOM} --chain-id $CHAIN_ID
    fi
}

# Adds all the non-validator accounts
add_accounts() {
    echo "Adding admin account..."
    admin_key=$(cat $KEYS_FILE | jq -c '.admin')
    add_genesis_account "$admin_key" ${USER_BALANCE}${DENOM} 

    echo "Adding faucet account..."
    faucet_key=$(cat $KEYS_FILE | jq -c '.faucet')
    add_genesis_account "$faucet_key" ${USER_BALANCE}${DENOM} 

    echo "Adding relayer accounts..."
    jq -c '.relayers[]' $KEYS_FILE | while IFS= read -r relayer_key; do
        add_genesis_account "$relayer_key" ${RELAYER_BALANCE}${DENOM}
    done

    echo "Adding user accounts..."
    jq -c '.users[]' $KEYS_FILE | while IFS= read -r user_keys; do
        add_genesis_account "$user_keys" ${RELAYER_BALANCE}${DENOM}
    done
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
    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="mint") ).duration |= "'$STRIDE_MINT_EPOCH_DURATION'"' $genesis_json 

    # Seed POA genesis with the validator set assembled in add_validators(). Post-v33,
    # ccvconsumer is no longer in the module manager and POA is the sole source of the
    # InitChain validator set. Bech32 below is authtypes.NewModuleAddress("gov") for
    # Stride's address prefix; see app/test_setup.go for the Go derivation.
    POA_ADMIN="stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl"
    jq --arg admin "$POA_ADMIN" \
       --argjson vals "$POA_VALIDATORS_JSON" \
       '.app_state.poa.params.admin = $admin
        | .app_state.poa.validators = $vals' \
       $genesis_json > /tmp/$(basename $genesis_json) && mv /tmp/$(basename $genesis_json) $genesis_json

    jq_inplace '.app_state.airdrop.params.period_length_seconds |= "'${AIRDROP_PERIOD_SECONDS}'"' $genesis_json 

    jq_inplace '.app_state.icqoracle.params.osmosis_chain_id |= "'$ICQORACLE_OSMOSIS_CHAIN_ID'"' $genesis_json
    jq_inplace '.app_state.icqoracle.params.osmosis_connection_id |= "'$ICQORACLE_OSMOSIS_CONNECTION_ID'"' $genesis_json
    jq_inplace '.app_state.icqoracle.params.update_interval_sec |= "'$ICQORACLE_UPDATE_INTERVAL_SEC'"' $genesis_json
    jq_inplace '.app_state.icqoracle.params.price_expiration_timeout_sec |= "'$ICQORACLE_PRICE_EXPIRATION_TIMEOUT_SEC'"' $genesis_json
}

# Genesis updates specific to non-stride chains
update_host_genesis() {
    echo "Updating genesis.json with host configuration..."

    if [[ "$CHAIN_NAME" == "osmosis" ]]; then
        strd_on_osmo="ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA"
        atom_on_osmo="ibc/6CDD4663F2F09CD62285E2D45891FC149A3568E316CE3EBBE201A71A78A69388" # through stride
        jq_inplace '.app_state.concentratedliquidity.params.is_permissionless_pool_creation_enabled |= true' $genesis_json 
    fi
}

# Saves the genesis file in the API
save_genesis() {
    echo "Saving genesis.json..."

    upload_shared_file $genesis_json ${GENESIS_DIR}/${CHAIN_NAME}/genesis.json
}

main() {
    echo "Initializing chain..."
    init_config
    add_validators
    add_accounts
    update_default_genesis
    if [[ "$CHAIN_NAME" == "stride" ]]; then 
        update_stride_genesis
    else 
        update_host_genesis
        $BINARY $chain_genesis_command collect-gentxs
    fi
    save_genesis
    echo "Done"
}

main >> logs/startup.log 2>&1 