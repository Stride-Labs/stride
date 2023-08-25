#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/../config.sh

CHAIN="$1"
KEYS_LOGS=$DOCKERNET_HOME/logs/keys.log

CHAIN_ID=$(GET_VAR_VALUE    ${CHAIN}_CHAIN_ID)
BINARY=$(GET_VAR_VALUE      ${CHAIN}_BINARY)
MAIN_CMD=$(GET_VAR_VALUE    ${CHAIN}_MAIN_CMD)
DENOM=$(GET_VAR_VALUE       ${CHAIN}_DENOM)
RPC_PORT=$(GET_VAR_VALUE    ${CHAIN}_RPC_PORT)
NUM_NODES=$(GET_VAR_VALUE   ${CHAIN}_NUM_NODES)
NODE_PREFIX=$(GET_VAR_VALUE ${CHAIN}_NODE_PREFIX)
VAL_PREFIX=$(GET_VAR_VALUE  ${CHAIN}_VAL_PREFIX)

# THe host zone can optionally specify additional the micro-denom granularity
# If they don't specify the ${CHAIN}_MICRO_DENOM_UNITS variable,
# EXTRA_MICRO_DENOM_UNITS will include 6 0's
MICRO_DENOM_UNITS_VAR_NAME=${CHAIN}_MICRO_DENOM_UNITS
MICRO_DENOM_UNITS="${!MICRO_DENOM_UNITS_VAR_NAME:-000000}"

VAL_TOKENS=${VAL_TOKENS}${MICRO_DENOM_UNITS}
STAKE_TOKENS=${STAKE_TOKENS}${MICRO_DENOM_UNITS}
ADMIN_TOKENS=${ADMIN_TOKENS}${MICRO_DENOM_UNITS}
USER_TOKENS=${USER_TOKENS}${MICRO_DENOM_UNITS}

set_stride_epochs() {
    genesis_config=$1

    # set epochs
    jq '(.app_state.epochs.epochs[] | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $STRIDE_DAY_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[] | select(.identifier=="hour") ).duration = $epochLen' --arg epochLen $STRIDE_HOUR_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[] | select(.identifier=="stride_epoch") ).duration = $epochLen' --arg epochLen $STRIDE_EPOCH_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[] | select(.identifier=="mint") ).duration = $epochLen' --arg epochLen $STRIDE_MINT_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
}

set_stride_genesis() {
    genesis_config=$1

    # set epochs
    set_stride_epochs $genesis_config

    # set params
    jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '.app_state.gov.params.max_deposit_period = $newVal' --arg newVal "$MAX_DEPOSIT_PERIOD" $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '.app_state.gov.params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $genesis_config > json.tmp && mv json.tmp $genesis_config

    # enable stride as an interchain accounts controller
    jq "del(.app_state.interchain_accounts)" $genesis_config > json.tmp && mv json.tmp $genesis_config
    interchain_accts=$(cat $DOCKERNET_HOME/config/ica_controller.json)
    jq ".app_state += $interchain_accts" $genesis_config > json.tmp && mv json.tmp $genesis_config
}

set_host_genesis() {
    genesis_config=$1

    # Shorten epochs and unbonding time
    jq '(.app_state.epochs.epochs[]? | select(.identifier=="day") ).duration = $epochLen' --arg epochLen $HOST_DAY_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[]? | select(.identifier=="hour") ).duration = $epochLen' --arg epochLen $HOST_HOUR_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[]? | select(.identifier=="week") ).duration = $epochLen' --arg epochLen $HOST_WEEK_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '(.app_state.epochs.epochs[]? | select(.identifier=="mint") ).duration = $epochLen' --arg epochLen $HOST_MINT_EPOCH_DURATION $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq '.app_state.gov.voting_params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $genesis_config > json.tmp && mv json.tmp $genesis_config

    # Set the mint start time to the genesis time if the chain configures inflation at the block level (e.g. stars)
    # also reduce the number of initial annual provisions so the inflation rate is not too high
    genesis_time=$(jq .genesis_time $genesis_config | tr -d '"')
    jq 'if .app_state.mint.params.start_time? then .app_state.mint.params.start_time=$newVal else . end' --arg newVal "$genesis_time" $genesis_config > json.tmp && mv json.tmp $genesis_config
    jq 'if .app_state.mint.params.initial_annual_provisions? then .app_state.mint.params.initial_annual_provisions=$newVal else . end' --arg newVal "$INITIAL_ANNUAL_PROVISIONS" $genesis_config > json.tmp && mv json.tmp $genesis_config

    # Add interchain accounts to the genesis set
    jq "del(.app_state.interchain_accounts)" $genesis_config > json.tmp && mv json.tmp $genesis_config
    interchain_accts=$(cat $DOCKERNET_HOME/config/ica_host.json)
    jq ".app_state += $interchain_accts" $genesis_config > json.tmp && mv json.tmp $genesis_config

    # Slightly harshen slashing parameters (if 5 blocks are missed, the validator will be slashed)
    # This makes it easier to test updating weights after a host zone validator is slashed
    sed -i -E 's|"signed_blocks_window": "100"|"signed_blocks_window": "10"|g' $genesis_config
    sed -i -E 's|"downtime_jail_duration": "600s"|"downtime_jail_duration": "10s"|g' $genesis_config
    sed -i -E 's|"slash_fraction_downtime": "0.010000000000000000"|"slash_fraction_downtime": "0.050000000000000000"|g' $genesis_config

    # LSM params
    if [[ "$CHAIN" == "GAIA" ]]; then
        jq '.app_state.staking.params.validator_bond_factor = $newVal' --arg newVal "$LSM_VALIDATOR_BOND_FACTOR" $genesis_config > json.tmp && mv json.tmp $genesis_config
        jq '.app_state.staking.params.validator_liquid_staking_cap = $newVal' --arg newVal "$LSM_VALIDATOR_LIQUID_STAKING_CAP" $genesis_config > json.tmp && mv json.tmp $genesis_config
        jq '.app_state.staking.params.global_liquid_staking_cap = $newVal' --arg newVal "$LSM_GLOBAL_LIQUID_STAKING_CAP" $genesis_config > json.tmp && mv json.tmp $genesis_config
    fi
}

set_consumer_genesis() {
    genesis_config=$1

    # add consumer genesis
    $MAIN_CMD add-consumer-section $NUM_NODES
    jq '.app_state.ccvconsumer.params.unbonding_period = $newVal' --arg newVal "$UNBONDING_TIME" $genesis_config > json.tmp && mv json.tmp $genesis_config
}

MAIN_ID=1 # Node responsible for genesis and persistent_peers
MAIN_NODE_NAME=""
MAIN_NODE_ID=""
MAIN_CONFIG=""
MAIN_GENESIS=""
echo "Initializing $CHAIN chain..."
for (( i=1; i <= $NUM_NODES; i++ )); do
    # Node names will be of the form: "stride1"
    node_name="${NODE_PREFIX}${i}"
    # Moniker is of the form: STRIDE_1
    moniker=$(printf "${NODE_PREFIX}_${i}" | awk '{ print toupper($0) }')

    # Create a state directory for the current node and initialize the chain
    mkdir -p $STATE/$node_name
    
    # If the chains commands are run only from docker, grab the command from the config
    # Otherwise, if they're run locally, append the home directory
    if [[ $BINARY == docker-compose* ]]; then
        cmd=$BINARY
    else
        cmd="$BINARY --home ${STATE}/$node_name"
    fi

    # Initialize the chain
    $cmd init $moniker --chain-id $CHAIN_ID --overwrite &> /dev/null
    chmod -R 777 $STATE/$node_name

    # Update node networking configuration 
    config_toml="${STATE}/${node_name}/config/config.toml"
    client_toml="${STATE}/${node_name}/config/client.toml"
    app_toml="${STATE}/${node_name}/config/app.toml"
    genesis_json="${STATE}/${node_name}/config/genesis.json"

    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" $config_toml
    sed -i -E "s|127.0.0.1|0.0.0.0|g" $config_toml
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" $config_toml
    sed -i -E "s|prometheus = false|prometheus = true|g" $config_toml

    sed -i -E "s|minimum-gas-prices = \".*\"|minimum-gas-prices = \"0${DENOM}\"|g" $app_toml
    sed -i -E '/\[api\]/,/^enable = .*$/ s/^enable = .*$/enable = true/' $app_toml
    sed -i -E 's|unsafe-cors = .*|unsafe-cors = true|g' $app_toml
    sed -i -E "s|snapshot-interval = 0|snapshot-interval = 300|g" $app_toml
    sed -i -E 's|localhost|0.0.0.0|g' $app_toml

    sed -i -E "s|chain-id = \"\"|chain-id = \"${CHAIN_ID}\"|g" $client_toml
    sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $client_toml
    sed -i -E "s|node = \".*\"|node = \"tcp://localhost:$RPC_PORT\"|g" $client_toml

    sed -i -E "s|\"stake\"|\"${DENOM}\"|g" $genesis_json 
    sed -i -E "s|\"aphoton\"|\"${DENOM}\"|g" $genesis_json # ethermint default

    # add a validator account
    val_acct="${VAL_PREFIX}${i}"
    val_mnemonic="${VAL_MNEMONICS[((i-1))]}"
    echo "$val_mnemonic" | $cmd keys add $val_acct --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
    val_addr=$($cmd keys show $val_acct --keyring-backend test -a | tr -cd '[:alnum:]._-')
    # Add this account to the current node
    $cmd add-genesis-account ${val_addr} ${VAL_TOKENS}${DENOM}

    # Copy over the provider stride validator keys to the provider (in the event
    # that we are testing ICS)
    if [[ $CHAIN == "GAIA" && -d $DOCKERNET_HOME/state/${STRIDE_NODE_PREFIX}${i} ]]; then
        stride_config=$DOCKERNET_HOME/state/${STRIDE_NODE_PREFIX}${i}/config
        host_config=$DOCKERNET_HOME/state/${NODE_PREFIX}${i}/config
        cp ${stride_config}/priv_validator_key.json ${host_config}/priv_validator_key.json
        cp ${stride_config}/node_key.json ${host_config}/node_key.json
    fi

    # Only generate the validator txs for host chains
    if [[ "$CHAIN" != "STRIDE" && "$CHAIN" != "HOST" ]]; then 
        $cmd gentx $val_acct ${STAKE_TOKENS}${DENOM} --chain-id $CHAIN_ID --keyring-backend test &> /dev/null
    fi
    
    # Get the endpoint and node ID
    node_id=$($cmd tendermint show-node-id)@$node_name:$PEER_PORT
    echo "Node #$i ID: $node_id"

    # Cleanup from seds
    rm -rf ${client_toml}-E
    rm -rf ${genesis_json}-E
    rm -rf ${app_toml}-E

    if [ $i -eq $MAIN_ID ]; then
        MAIN_NODE_NAME=$node_name
        MAIN_NODE_ID=$node_id
        MAIN_CONFIG=$config_toml
        MAIN_GENESIS=$genesis_json
    else
        # also add this account and it's genesis tx to the main node
        $MAIN_CMD add-genesis-account ${val_addr} ${VAL_TOKENS}${DENOM}
        if [ -d "${STATE}/${node_name}/config/gentx" ]; then
            cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${MAIN_NODE_NAME}/config/gentx/
        fi

        # and add each validator's keys to the first state directory
        echo "$val_mnemonic" | $MAIN_CMD keys add $val_acct --recover --keyring-backend=test &> /dev/null
    fi
done

if [ "$CHAIN" == "STRIDE" ]; then 
    # add the stride admin account
    echo "$STRIDE_ADMIN_MNEMONIC" | $MAIN_CMD keys add $STRIDE_ADMIN_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
    STRIDE_ADMIN_ADDRESS=$($MAIN_CMD keys show $STRIDE_ADMIN_ACCT --keyring-backend test -a)
    $MAIN_CMD add-genesis-account ${STRIDE_ADMIN_ADDRESS} ${ADMIN_TOKENS}${DENOM}

    # add relayer accounts
    for i in "${!RELAYER_ACCTS[@]}"; do
        RELAYER_ACCT="${RELAYER_ACCTS[i]}"
        RELAYER_MNEMONIC="${RELAYER_MNEMONICS[i]}"

        echo "$RELAYER_MNEMONIC" | $MAIN_CMD keys add $RELAYER_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
        
        RELAYER_ADDRESS=$($MAIN_CMD keys show $RELAYER_ACCT --keyring-backend test -a)
        $MAIN_CMD add-genesis-account ${RELAYER_ADDRESS} ${VAL_TOKENS}${DENOM}
    done
else 
    # add a revenue account
    REV_ACCT_VAR=${CHAIN}_REV_ACCT
    REV_ACCT=${!REV_ACCT_VAR}
    echo $REV_MNEMONIC | $MAIN_CMD keys add $REV_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1

    # add a relayer account
    RELAYER_ACCT=$(GET_VAR_VALUE     RELAYER_${CHAIN}_ACCT)
    RELAYER_MNEMONIC=$(GET_VAR_VALUE RELAYER_${CHAIN}_MNEMONIC)

    echo "$RELAYER_MNEMONIC" | $MAIN_CMD keys add $RELAYER_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
    RELAYER_ADDRESS=$($MAIN_CMD keys show $RELAYER_ACCT --keyring-backend test -a | tr -cd '[:alnum:]._-')
    $MAIN_CMD add-genesis-account ${RELAYER_ADDRESS} ${VAL_TOKENS}${DENOM}

    if [ "$CHAIN" == "GAIA" ]; then 
        echo "$RELAYER_GAIA_ICS_MNEMONIC" | $MAIN_CMD keys add $RELAYER_GAIA_ICS_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
        RELAYER_ADDRESS=$($MAIN_CMD keys show $RELAYER_GAIA_ICS_ACCT --keyring-backend test -a | tr -cd '[:alnum:]._-')
        $MAIN_CMD add-genesis-account ${RELAYER_ADDRESS} ${VAL_TOKENS}${DENOM}
    fi
fi

# add a staker account for integration tests
# the account should live on both stride and the host chain
echo "$USER_MNEMONIC" | $MAIN_CMD keys add $USER_ACCT --recover --keyring-backend=test >> $KEYS_LOGS 2>&1
USER_ADDRESS=$($MAIN_CMD keys show $USER_ACCT --keyring-backend test -a)
$MAIN_CMD add-genesis-account ${USER_ADDRESS} ${USER_TOKENS}${DENOM}

# Only collect the validator genesis txs for host chains
if [[ "$CHAIN" != "STRIDE" && "$CHAIN" != "HOST" ]]; then 
    # now we process gentx txs on the main node
    $MAIN_CMD collect-gentxs &> /dev/null
fi

# wipe out the persistent peers for the main node (these are incorrectly autogenerated for each validator during collect-gentxs)
sed -i -E "s|persistent_peers = .*|persistent_peers = \"\"|g" $MAIN_CONFIG

# update chain-specific genesis settings
if [ "$CHAIN" == "STRIDE" ]; then 
    set_stride_genesis $MAIN_GENESIS
else
    set_host_genesis $MAIN_GENESIS
fi

# update consumer genesis for stride binary chains
if [[ "$CHAIN" == "STRIDE" || "$CHAIN" == "HOST" ]]; then
    set_consumer_genesis $MAIN_GENESIS
fi

# the HOST chain must set the epochs to fulfill the invariant that the stride epoch
# is 1/4th the day epoch
if [[ "$CHAIN" == "HOST" ]]; then
    set_stride_epochs $MAIN_GENESIS
fi

# for all peer nodes....
for (( i=2; i <= $NUM_NODES; i++ )); do
    node_name="${NODE_PREFIX}${i}"
    config_toml="${STATE}/${node_name}/config/config.toml"
    genesis_json="${STATE}/${node_name}/config/genesis.json"

    # add the main node as a persistent peer
    sed -i -E "s|persistent_peers = .*|persistent_peers = \"${MAIN_NODE_ID}\"|g" $config_toml
    # copy the main node's genesis to the peer nodes to ensure they all have the same genesis
    cp $MAIN_GENESIS $genesis_json

    rm -rf ${config_toml}-E
done

# Cleanup from seds
rm -rf ${MAIN_CONFIG}-E
rm -rf ${MAIN_GENESIS}-E