# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for chain_name in ${STRIDE_CHAINS[@]}; do
    mkdir -p ./$STATE/$chain_name
done

# fetch the stride node ids
node1=$($ST1_RUN tendermint show-node-id)@stride1:26656
node2=$($ST2_RUN tendermint show-node-id)@stride2:26656
node3=$($ST3_RUN tendermint show-node-id)@stride3:26656
echo ${node1}
STRIDE_NODES=($node1 $node2 $node3)

# then, we initialize our chains 
echo 'Initializing chains...'
for i in "${!STRIDE_CHAINS[@]}"; do
    chain_name=${STRIDE_CHAINS[i]}
    vkey=${VKEYS[i]}
    val_acct=${VAL_ACCTS[i]}
    echo "\t$chain_name"
    $BASE_RUN init test --chain-id $chain_name --overwrite --home "$STATE/$chain_name" 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${chain_name}/config/genesis.json"
    # add VALidator account
    echo $vkey | $BASE_RUN keys add $val_acct --recover --keyring-backend=test --home "$STATE/$chain_name" > /dev/null
    # get validator address
    VAL_ADDR=$($BASE_RUN keys show $val_acct --keyring-backend test -a --home "$STATE/$chain_name")
    # add money for this validator account
    # $BASE_RUN add-genesis-account ${VAL_ADDR} 500000000000ustrd --home "$STATE/$main_chain"
    $BASE_RUN add-genesis-account ${VAL_ADDR} 500000000000ustrd --home "$STATE/$chain_name"
    # actually set this account as a validator
    yes | $BASE_RUN gentx $val_acct 1000000000ustrd --chain-id $chain_name --keyring-backend test --home "$STATE/$chain_name"
    # this is just annoying, but we need to move the validator tx
    # cp "./${STATE}/${chain_name}/config/gentx/*.json" "./${STATE}/${chain_name}/config/gentx/"
    # now we process these txs 
    $BASE_RUN collect-gentxs --home "$STATE/$chain_name" 2> /dev/null
    # add peers in config.toml so that nodes can find each other by constructing a fully connected
    # graph of nodes
    # sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${chain_name}/config/genesis.json"
    peers=""
    for j in "${!STRIDE_NODES[@]}"; do
        if [ $j -ne $i ]
        then
            # peers+=(${STRIDE_NODES[j]})
            peers="${STRIDE_NODES[j]},${peers}"
        fi
    done
    echo 'peers are: '
    echo $peers
    sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$peers\"|g" "${STATE}/${chain_name}/config/config.toml"
done


# make sure all Stride chains have the same genesis 
for i in "${!STRIDE_CHAINS[@]}"; do
    cp ./${STATE}/${main_chain}/config/genesis.json ./${STATE}/${STRIDE_CHAINS[i]}/config/genesis.json
done
# strided start --home state/STRIDE_1  # TESTING ONLY
# next we build our docker images
docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .  # builds from scratch
# docker build --tag stridezone:stride -f Dockerfile.stride .  # uses cache to speed things up

# finally we serve our docker images
sleep 5
docker-compose up -d stride1 stride2 stride3


