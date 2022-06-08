# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE/gaia*
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for node_name in ${GAIA_DOCKER_NAMES[@]}; do
    mkdir -p $STATE/$node_name
done

# fetch the stride node ids
GAIA_NODES=()
# then, we initialize our chains 
echo 'Initializing chains...'

for i in ${!GAIA_CHAINS[@]}; do
    chain_name=${GAIA_CHAINS[i]}
    node_name=${GAIA_DOCKER_NAMES[i]}
    vkey=${GVKEYS[i]}
    val_acct=${GVAL_ACCTS[i]}
    gaia_cmd=${GAIA_CMDS[i]}
    echo "\t$node_name"
    $gaia_cmd init test --chain-id $chain_name --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${node_name}/config/genesis.json"
    sed -i -E 's|"full"|"validator"|g' "${STATE}/${node_name}/config/config.toml"
    # add VALidator account
    echo $vkey | $gaia_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    VAL_ADDR=$($gaia_cmd keys show $val_acct --keyring-backend test -a) > /dev/null
    # add money for this validator account
    $gaia_cmd add-genesis-account ${VAL_ADDR} 500000000000uatom
    # actually set this account as a validator
    yes | $gaia_cmd gentx $val_acct 1000000000uatom --chain-id $main_gaia_chain --keyring-backend test 2> /dev/null
    # now we process these txs 
    $gaia_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    dock_name=${GAIA_DOCKER_NAMES[i]}
    node_id=$($gaia_cmd tendermint show-node-id)@$dock_name:$PORT_ID
    GAIA_NODES+=( $node_id )

    if [ $i -ne $MAIN_ID ]; then
        $main_gaia_cmd add-genesis-account ${VAL_ADDR} 500000000000uatom
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${main_gaia_node}/config/gentx/
    fi
done

# Restore relayer account on gaia
echo $RLY_MNEMONIC_2 | $main_gaia_cmd keys add rly2 --recover --keyring-backend=test > /dev/null
RLY_ADDRESS_2=$($main_gaia_cmd keys show rly2 --keyring-backend test -a)
# Give relayer account token balance
$main_gaia_cmd add-genesis-account ${RLY_ADDRESS_2} 500000000000ustrd

$main_gaia_cmd collect-gentxs 2> /dev/null

# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!GAIA_DOCKER_NAMES[@]}; do
    node_name=${GAIA_DOCKER_NAMES[i]}
    peers=""
    for j in "${!GAIA_CHAINS[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${GAIA_NODES[j]},${peers}"
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
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${STATE}/${main_gaia_node}/config/genesis.json"


# make sure all Gaia chains have the same genesis
for i in "${!GAIA_CHAINS[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${main_gaia_node}/config/genesis.json ${STATE}/${GAIA_DOCKER_NAMES[i]}/config/genesis.json
    fi
done

# docker compose up -d gaia1 gaia2 gaia3