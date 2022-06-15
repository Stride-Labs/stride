# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE/gaia*
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for node_name in ${GAIA_NODE_NAMES[@]}; do
    mkdir -p $STATE/$node_name
done

# fetch the stride node ids
GAIA_NODE_IDS=()
# then, we initialize our chains 
echo 'Initializing chains...'

for i in ${!GAIA_NODE_NAMES[@]}; do
    node_name=${GAIA_NODE_NAMES[i]}
    vkey=${GAIA_VAL_KEYS[i]}
    val_acct=${GAIA_VAL_ACCTS[i]}
    gaia_cmd=${GAIA_CMDS[i]}
    echo "\t$node_name"
    $gaia_cmd init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${node_name}/config/genesis.json"
    sed -i -E 's|"full"|"validator"|g' "${STATE}/${node_name}/config/config.toml"
    # add validator account
    echo $vkey | $gaia_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    val_addr=$($gaia_cmd keys show $val_acct --keyring-backend test -a) > /dev/null
    # add money for this validator account
    $gaia_cmd add-genesis-account ${val_addr} 500000000000000uatom
    # actually set this account as a validator
    $gaia_cmd gentx $val_acct 1000000000uatom --chain-id $GAIA_CHAIN --keyring-backend test 2> /dev/null
    # now we process these txs 
    $gaia_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    node_id=$($gaia_cmd tendermint show-node-id)@$node_name:$PORT_ID
    GAIA_NODE_IDS+=( $node_id )

    if [ $i -ne $MAIN_ID ]; then
        $GAIA_MAIN_CMD add-genesis-account ${val_addr} 50000000000000uatom
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${GAIA_MAIN_NODE}/config/gentx/
    fi
done

# Restore relayer account on gaia
echo $RLY_MNEMONIC_2 | $GAIA_MAIN_CMD keys add rly2 --recover --keyring-backend=test > /dev/null
RLY_ADDRESS_2=$($GAIA_MAIN_CMD keys show rly2 --keyring-backend test -a)
# Give relayer account token balance
$GAIA_MAIN_CMD add-genesis-account ${RLY_ADDRESS_2} 5000000000000ustrd

$GAIA_MAIN_CMD collect-gentxs 2> /dev/null

# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!GAIA_NODE_NAMES[@]}; do
    node_name=${GAIA_NODE_NAMES[i]}
    peers=""
    for j in "${!GAIA_NODE_NAMES[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${GAIA_NODE_IDS[j]},${peers}"
        fi
    done
    sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$peers\"|g" "${STATE}/${node_name}/config/config.toml"
    # use blind address (not loopback) to allow incoming connections from outside networks for local debugging
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${node_name}/config/config.toml"
    sed -i -E "s|minimum-gas-prices = \"\"|minimum-gas-prices = \"0uatom\"|g" "${STATE}/${node_name}/config/app.toml"
    # allow CORS and API endpoints for block explorer
    sed -i -E 's|enable = false|enable = true|g' "${STATE}/${node_name}/config/app.toml"
    sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' "${STATE}/${node_name}/config/app.toml"
done

## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${GAIA_MAIN_NODE}/config/genesis.json"


# make sure all Gaia nodes have the same genesis
for i in "${!GAIA_NODE_NAMES[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${GAIA_MAIN_NODE}/config/genesis.json ${STATE}/${GAIA_NODE_NAMES[i]}/config/genesis.json
    fi
done
