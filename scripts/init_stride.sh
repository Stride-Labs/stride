#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

docker compose down

# first, we need to create some saved state, so that we can copy to docker files
for node_name in ${STRIDE_NODE_NAMES[@]}; do
    mkdir -p $STATE/$node_name
done

# fetch the stride node ids
STRIDE_NODE_IDS=()
# then, we initialize our chains 
echo 'Initializing chains...'
for i in ${!STRIDE_NODE_NAMES[@]}; do
    node_name=${STRIDE_NODE_NAMES[i]}
    vkey=${STRIDE_VAL_KEYS[i]}
    val_acct=${STRIDE_VAL_ACCTS[i]}
    st_cmd=${STRIDE_CMDS[i]}
    echo "\t$node_name"
    $st_cmd init test --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${node_name}/config/genesis.json"
    # add validator account
    echo $vkey | $st_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    val_addr=$($st_cmd keys show $val_acct --keyring-backend test -a)
    # add money for this validator account
    $st_cmd add-genesis-account ${val_addr} 500000000000ustrd
    # actually set this account as a validator
    $st_cmd gentx $val_acct 1000000000ustrd --chain-id $STRIDE_CHAIN --keyring-backend test 2> /dev/null
    # now we process these txs 
    $st_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    node_id=$($st_cmd tendermint show-node-id)@$node_name:$PORT_ID
    STRIDE_NODE_IDS+=( $node_id )

    if [ $i -ne $MAIN_ID ]
    then
        $STRIDE_MAIN_CMD add-genesis-account ${val_addr} 500000000000ustrd
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${STRIDE_MAIN_NODE}/config/gentx/
    fi
done

# modify Stride epoch to be 3s
main_config=$STATE/${STRIDE_MAIN_NODE}/config/genesis.json
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "3s" $main_config > json.tmp && mv json.tmp $main_config

# Restore relayer account on stride
echo $RLY_MNEMONIC_1 | $STRIDE_MAIN_CMD keys add rly1 --recover --keyring-backend=test > /dev/null
RLY_ADDRESS_1=$($STRIDE_MAIN_CMD keys show rly1 --keyring-backend test -a)
# Give relayer account token balance
$STRIDE_MAIN_CMD add-genesis-account ${RLY_ADDRESS_1} 500000000000ustrd

$STRIDE_MAIN_CMD collect-gentxs 2> /dev/null
# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!STRIDE_NODE_NAMES[@]}; do
    node_name=${STRIDE_NODE_NAMES[i]}
    peers=""
    for j in "${!STRIDE_NODE_IDS[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${STRIDE_NODE_IDS[j]},${peers}"
        fi
    done
    sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$peers\"|g" "${STATE}/${node_name}/config/config.toml"
    # use blind address (not loopback) to allow incoming connections from outside networks for local debugging
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${node_name}/config/config.toml"
done

# make sure all Stride nodes have the same genesis
for i in "${!STRIDE_NODE_NAMES[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${STRIDE_MAIN_NODE}/config/genesis.json ${STATE}/${STRIDE_NODE_NAMES[i]}/config/genesis.json
    fi
done