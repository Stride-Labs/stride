# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE/GAIA*
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for chain_name in ${GAIA_CHAINS[@]}; do
    mkdir -p $STATE/$chain_name
done

# fetch the stride node ids
GAIA_NODES=()
# then, we initialize our chains 
echo 'Initializing chains...'

for i in ${!GAIA_CHAINS[@]}; do
    chain_name=${GAIA_CHAINS[i]}
    vkey=${GVKEYS[i]}
    val_acct=${GVAL_ACCTS[i]}
    gaia_cmd=${GAIA_CMDS[i]}
    echo "\t$chain_name"
    $gaia_cmd init test --chain-id $chain_name --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"uatom"|g' "${STATE}/${chain_name}/config/genesis.json"
    # add VALidator account
    echo $vkey | $gaia_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    VAL_ADDR=$($gaia_cmd keys show $val_acct --keyring-backend test -a) > /dev/null
    # add money for this validator account
    $gaia_cmd add-genesis-account ${VAL_ADDR} 500000000000uatom
    # actually set this account as a validator
    yes | $gaia_cmd gentx $val_acct 1000000000uatom --chain-id $main_gaia_chain --keyring-backend test
    # now we process these txs 
    $gaia_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    dock_name=${GAIA_DOCKER_NAMES[i]}
    node_id=$($gaia_cmd tendermint show-node-id)@$dock_name:$PORT_ID
    GAIA_NODES+=( $node_id )
# gaiad --home /gaia/.gaiad keys show gval1 --keyring-backend test -a
# gaiad --home /gaia/.gaiad gentx gval1 1000000000uatom --chain-id GAIA_1 --keyring-backend test
# gaiad --home /gaia/.gaiad collect-gentxs
# gaiad start --home /gaia/.gaiad
# sed -i -E "s|minimum-gas-prices = \"\"|minimum-gas-prices = \"0uatom\"|g" "/gaia/.gaiad/config/app.toml"

    # if [ $i -ne $MAIN_ID ]; then
    #     $main_gaia_cmd add-genesis-account ${VAL_ADDR} 500000000000uatom
     #    cp ${STATE}/${chain_name}/config/gentx/*.json ${STATE}/${main_gaia_chain}/config/gentx/
    # fi
done

$main_gaia_cmd collect-gentxs 2> /dev/null

# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!GAIA_CHAINS[@]}; do
    chain_name=${GAIA_CHAINS[i]}
    peers=""
    for j in "${!GAIA_CHAINS[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${GAIA_NODES[j]},${peers}"
        fi
    done
    echo "${chain_name} peers are:"
    echo $peers
    # sed -i -E "s|persistent-peers = \"\"|persistent-peers = \"$peers\"|g" "${STATE}/${chain_name}/config/config.toml"
    # use blind address (not loopback) to allow incoming connections from outside networks for local debugging
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${chain_name}/config/config.toml"
    sed -i -E "s|minimum-gas-prices = \"\"|minimum-gas-prices = \"0uatom\"|g" "${STATE}/${chain_name}/config/app.toml"
done

# make sure all Stride chains have the same genesis
for i in "${!GAIA_CHAINS[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${main_gaia_chain}/config/genesis.json ${STATE}/${GAIA_CHAINS[i]}/config/genesis.json
    fi
done


# # finally we serve our docker images
sleep 5
# docker-compose down
# docker-compose up -d gaia1 gaia2 gaia3


