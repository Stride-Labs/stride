#!/bin/bash

set -eu

DEPLOYMENT_NAME="$1" # e.g. testnet
NETWORK_NAME="$2"    # e.g. stride
NUM_NODES="$3"    

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state
VAL_PREFIX=val
PORT_ID=26656
STRIDE_CMD=build/strided
VAL_TOKENS=500000000ustrd
STAKE_TOKENS=300000000ustrd

echo "Cleaning state"
rm -rf $STATE
mkdir $STATE
touch $STATE/keys.txt

# Initialize the chain, keeping track of node ids
PEER_NODE_IDS=""
SEED_NODE_ID=""
MAIN_NODE_CMD=""
SEED_ID=0
MAIN_ID=1 # Node responsible for genesis
echo 'Initializing state for each node in the chain...'
for (( i=0; i <= $num_nodes; i++ )); do
    # Seed nodes will be of the form: "stride-seed"
    # Val nodes will be of the form: "stride-node1"
    if [ $i -eq $SEED_ID ]; then 
        node_name="${NETWORK_NAME}-seed"
    else 
        node_name="${NETWORK_NAME}-node${i}"
    fi

    # Moniker is of the form: STRIDE_1
    moniker="${NETWORK_NAME^^}_${i}"
    # Create state directory for node and initialize chain
    mkdir -p $STATE/$node_name
    st_cmd="$STRIDE_CMD --home ${STATE}/$node_name"
    $st_cmd init $moniker --chain-id $NETWORK_NAME --overwrite 2> /dev/null

    # Update node networking configuration 
    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" "${STATE}/${node_name}/config/config.toml"
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${node_name}/config/config.toml"

    # Get the endpoint and node ID
    endpoint="${node_name}.${DEPLOYMENT_NAME}.stridelabs.co"
    node_id=$($st_cmd tendermint show-node-id)@$endpoint:$PORT_ID
    echo "Node ID: $node_id"

    if [ $i -eq $SEED_ID ]; then
        # If it's a seed node, update the config to indicate seed_mode
        sed -i -E 's|seed_mode = false|seed_mode = true|g' "${STATE}/${seed_node}/config/config.toml"
        SEED_NODE_ID=$node_id
    else if [ $i -eq $MAIN_ID ]; then
        # If it's the main node, update the denom in the genesis file and store the main command
        sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${node_name}/config/genesis.json"
        MAIN_NODE_CMD=$st_cmd
    else
        # add validator account
        val_acct="${VAL_PREFIX}${i}"
        $st_cmd keys add $val_acct --keyring-backend=test >> $STATE/keys.txt 2>&1
        val_addr=$($st_cmd keys show $val_acct --keyring-backend test -a)
        # Add this account to both the current node and the main node
        $st_cmd add-genesis-account ${val_addr} $VAL_TOKENS
        $MAIN_NODE_CMD add-genesis-account ${val_addr} $VAL_TOKENS
        # actually set this account as a validator on the current node and copy that tx back to the main node
        $st_cmd gentx $val_acct $STAKE_TOKENS --chain-id $NETWORK_NAME --keyring-backend test 2> /dev/null
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${main_node}/config/gentx/
        # add this node's id to the list of peer nodes that will be used by the seed node
        PEER_NODE_IDS="${node_id},${PEER_NODE_IDS}" 
        # set the seed node as the only peer for this validator 
        sed -i -E "s|persistent_peers = .*|persistent_peers = \"$PEER_NODE_IDS\"|g" "${STATE}/${NETWORK_NAME}-seed/config/config.toml"
    fi
done

# now we process gentx txs on the main node
$MAIN_NODE_CMD collect-gentxs 2> /dev/null

# add peer nodes to the seed node's config so that nodes can find each other 
sed -i -E "s|persistent_peers = .*|persistent_peers = \"$PEER_NODE_IDS\"|g" "${STATE}/${NETWORK_NAME}-seed/config/config.toml"

# copy the main node's genesis to the other nodes to ensure all nodes have the same genesis
for (( i=0; i <= $num_nodes; i++ )); do
    if [ $i -ne $MAIN_ID ]; then
        node_name="${NETWORK_NAME}-node${i}"
        cp ${STATE}/${main_node}/config/genesis.json ${STATE}/${node_name}/config/genesis.json
    fi
done