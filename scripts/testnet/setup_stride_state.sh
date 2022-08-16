#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

NETWORK_NAME=stride
VAL_PREFIX=val
VAL_TOKENS=5000000000000ustrd
STAKE_TOKENS=3000000000000ustrd
FAUCET_TOKENS=10000000000000000ustrd
STRIDE_ADMIN_ACCT=stride
STRIDE_ADMIN_TOKENS=1000000000ustrd
NUM_NODES=$NUM_STRIDE_NODES

DAY_EPOCH_DURATION="3600s"
STRIDE_EPOCH_DURATION="600s"
UNBONDING_TIME="21600s"
MAX_DEPOSIT_PERIOD="3600s"
VOTING_PERIOD="900s"
SIGNED_BLOCKS_WINDOW="30000"
MIN_SIGNED_PER_WINDOW="0.050000000000000000"
SLASH_FRACTION_DOWNTIME="0.001000000000000000"

PEER_NODE_IDS=""
MAIN_ID=1 # Node responsible for genesis and persistent_peers
MAIN_NODE_NAME=""
MAIN_NODE_CMD=""
MAIN_NODE_ID=""
MAIN_NODE_ENDPOINT=""
echo 'Initializing stride...'
for (( i=1; i <= $NUM_NODES; i++ )); do
    # Node names will be of the form: "stride-node1"
    node_name="${NETWORK_NAME}-node${i}"
    # Moniker is of the form: STRIDE_1
    moniker=$(printf "${NETWORK_NAME}_${i}" | awk '{ print toupper($0) }')

    # figure out how many tokens to give this account
    if [ $i -eq 2 ]; then
        NODE_TOKENS=$FAUCET_TOKENS
    else
        NODE_TOKENS=$VAL_TOKENS
    fi 
    # Create a state directory for the current node and initialize the chain
    mkdir -p $STATE/$node_name
    st_cmd="$STRIDE_CMD --home ${STATE}/$node_name"
    $st_cmd init $moniker --chain-id $STRIDE_CHAIN_ID --overwrite 2> /dev/null

    # Update node networking configuration 
    configtoml="${STATE}/${node_name}/config/config.toml"
    clienttoml="${STATE}/${node_name}/config/client.toml"
    apptoml="${STATE}/${node_name}/config/app.toml"
    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" $configtoml
    sed -i -E "s|127.0.0.1|0.0.0.0|g" $configtoml
    sed -i -E "s|timeout_commit = \"5s\"|timeout_commit = \"${BLOCK_TIME}\"|g" $configtoml
    sed -i -E "s|chain-id = \"\"|chain-id = \"${STRIDE_CHAIN_ID}\"|g" $clienttoml
    sed -i -E "s|keyring-backend = \"os\"|keyring-backend = \"test\"|g" $clienttoml
    # Add cert file
    # sed -i -E "s|tls_cert_file = \"\"|tls_cert_file = \"/stride/certfile.pem\"|g" $configtoml
    # sed -i -E "s|tls_key_file = \"\"|tls_key_file = \"/stride/certkey.pem\"|g" $configtoml
    # sed -i -E "s|localhost|127.0.0.1|g" $configtoml
    # sed -i -E "s|localhost|127.0.0.1|g" $clienttoml
    # Enable prometheus
    sed -i -E "s|prometheus = false|prometheus = true|g" $configtoml

    # update the denom in the genesis file 
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${node_name}/config/genesis.json"

    # Get the endpoint and node ID
    endpoint="${node_name}.${DEPLOYMENT_NAME}.${DOMAIN}"
    node_id=$($st_cmd tendermint show-node-id)@$endpoint:$PORT_ID
    echo "Node ID: $node_id"

    # Configure an NGINX reverse proxy
    nginx_conf="${STATE}/${node_name}/config/nginx.conf"
    cp ${SCRIPT_DIR}/nginx_config_template.conf $nginx_conf
    sed -i -E "s|HOME_DIR|stride|g" $nginx_conf
    sed -i -E "s|ENDPOINT|$endpoint|g" $nginx_conf
    rm -f "${nginx_conf}-e"

    # add a validator account
    val_acct="${VAL_PREFIX}${i}"
    $st_cmd keys add $val_acct --keyring-backend=test >> $STATE/keys.txt 2>&1
    val_addr=$($st_cmd keys show $val_acct --keyring-backend test -a)
    # Add this account to the current node
    $st_cmd add-genesis-account ${val_addr} $NODE_TOKENS
    # actually set this account as a validator on the current node 
    $st_cmd gentx $val_acct $STAKE_TOKENS --chain-id $STRIDE_CHAIN_ID --keyring-backend test 2> /dev/null
    
    # modify our snapshot interval
    sed -i -E "s|snapshot-interval = 0|snapshot-interval = 1000|g" $apptoml

    if [ $i -eq $MAIN_ID ]; then
        MAIN_NODE_NAME=$node_name
        MAIN_NODE_CMD=$st_cmd
        MAIN_NODE_ID=$node_id
        MAIN_NODE_ENDPOINT=$endpoint
    else
        # also add this account and it's genesis tx to the main node
        $MAIN_NODE_CMD add-genesis-account ${val_addr} $NODE_TOKENS
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${MAIN_NODE_NAME}/config/gentx/
    fi
done

# add Hermes and ICQ relayer accounts on Stride
$MAIN_NODE_CMD keys add $HERMES_STRIDE_ACCT --keyring-backend=test >> $STATE/keys.txt 2>&1
$MAIN_NODE_CMD keys add $ICQ_STRIDE_ACCT --keyring-backend=test >> $STATE/keys.txt 2>&1
HERMES_STRIDE_ADDRESS=$($MAIN_NODE_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
ICQ_STRIDE_ADDRESS=$($MAIN_NODE_CMD keys show $ICQ_STRIDE_ACCT --keyring-backend test -a)

# give relayer accounts token balances
$MAIN_NODE_CMD add-genesis-account ${HERMES_STRIDE_ADDRESS} $VAL_TOKENS
$MAIN_NODE_CMD add-genesis-account ${ICQ_STRIDE_ADDRESS} $VAL_TOKENS

source ${SCRIPT_DIR}/genesis.sh 

# Add the stride admin account
echo "$STRIDE_ADMIN_MNEMONIC" | $MAIN_NODE_CMD keys add $STRIDE_ADMIN_ACCT --recover --keyring-backend=test >> $STATE/keys.txt 2>&1
STRIDE_ADMIN_ADDRESS=$($MAIN_NODE_CMD keys show $STRIDE_ADMIN_ACCT --keyring-backend test -a)
$MAIN_NODE_CMD add-genesis-account ${STRIDE_ADMIN_ADDRESS} $STRIDE_ADMIN_TOKENS

# now we process gentx txs on the main node
$MAIN_NODE_CMD collect-gentxs 2> /dev/null

# wipe out the persistent peers for the main node (these are incorrectly autogenerated for each validator during collect-gentxs)
sed -i -E "s|persistent_peers = .*|persistent_peers = \"\"|g" "${STATE}/${MAIN_NODE_NAME}/config/config.toml"

# modify our stride epochs
main_genesis="${STATE}/${MAIN_NODE_NAME}/config/genesis.json"
jq '.app_state.epochs.epochs[1].duration = $newVal' --arg newVal "$DAY_EPOCH_DURATION" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "$STRIDE_EPOCH_DURATION" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.staking.params.unbonding_time = $newVal' --arg newVal "$UNBONDING_TIME" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.gov.deposit_params.max_deposit_period = $newVal' --arg newVal "$MAX_DEPOSIT_PERIOD" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.gov.voting_params.voting_period = $newVal' --arg newVal "$VOTING_PERIOD" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.slashing.params.signed_blocks_window = $newVal' --arg newVal "$SIGNED_BLOCKS_WINDOW" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.slashing.params.min_signed_per_window = $newVal' --arg newVal "$MIN_SIGNED_PER_WINDOW" $main_genesis > json.tmp && mv json.tmp $main_genesis
jq '.app_state.slashing.params.slash_fraction_downtime = $newVal' --arg newVal "$SLASH_FRACTION_DOWNTIME" $main_genesis > json.tmp && mv json.tmp $main_genesis

# for all peer nodes....
for (( i=2; i <= $NUM_NODES; i++ )); do
    node_name="${NETWORK_NAME}-node${i}"
    # add the main node as a persistent peer
    sed -i -E "s|persistent_peers = .*|persistent_peers = \"${MAIN_NODE_ID}\"|g" "${STATE}/${node_name}/config/config.toml"
    # copy the main node's genesis to the peer nodes to ensure they all have the same genesis
    cp $main_genesis ${STATE}/${node_name}/config/genesis.json
done

STRIDE_STARTUP_FILE="${STATE}/stride_startup.sh"
cp ${SCRIPT_DIR}/stride_startup_template.sh $STRIDE_STARTUP_FILE
sed -i -E "s|STRIDE_ENDPOINT|$MAIN_NODE_ENDPOINT|g" $STRIDE_STARTUP_FILE
