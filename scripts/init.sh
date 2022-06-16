# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for node_name in ${STRIDE_DOCKER_NAMES[@]}; do
    mkdir -p $STATE/$node_name
done

# run through init args, if needed
while getopts bdfsa flag; do
    case "${flag}" in
        b) ignite chain init ;;
        d) sh $SCRIPT_DIR/docker_build.sh ;;
        f) sh $SCRIPT_DIR/docker_build.sh -f ;;
        s) sh $SCRIPT_DIR/docker_build.sh -s ;;
        a) sh $SCRIPT_DIR/docker_build.sh -a ;;
    esac
done

# Init Stride
#############################################################################################################################
# fetch the stride node ids
STRIDE_NODES=()
# then, we initialize our chains 
echo 'Initializing chains...'
for i in ${!STRIDE_CHAINS[@]}; do
    chain_name=${STRIDE_CHAINS[i]}
    node_name=${STRIDE_DOCKER_NAMES[i]}
    vkey=${VKEYS[i]}
    val_acct=${VAL_ACCTS[i]}
    st_cmd=${ST_CMDS[i]}
    echo "\t$node_name"
    $st_cmd init test --chain-id $chain_name --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${node_name}/config/genesis.json"
    # add VALidator account
    echo $vkey | $st_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    VAL_ADDR=$($st_cmd keys show $val_acct --keyring-backend test -a)
    # add money for this validator account
    $st_cmd add-genesis-account ${VAL_ADDR} 500000000000ustrd
    # actually set this account as a validator
    yes | $st_cmd gentx $val_acct 1000000000ustrd --chain-id $main_chain --keyring-backend test 2> /dev/null
    # now we process these txs 
    $st_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    dock_name=${STRIDE_DOCKER_NAMES[i]}
    node_id=$($st_cmd tendermint show-node-id)@$dock_name:$PORT_ID
    STRIDE_NODES+=( $node_id )

    if [ $i -ne $MAIN_ID ]
    then
        $main_cmd add-genesis-account ${VAL_ADDR} 500000000000ustrd
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${main_node}/config/gentx/
    fi
done

# modify Stride epoch to be 3s
main_config=$STATE/${main_node}/config/genesis.json
jq '.app_state.epochs.epochs[2].duration = $newVal' --arg newVal "3s" $main_config > json.tmp && mv json.tmp $main_config

# Restore relayer account on stride
echo $RLY_MNEMONIC_1 | $main_cmd keys add rly1 --recover --keyring-backend=test > /dev/null
RLY_ADDRESS_1=$($main_cmd keys show rly1 --keyring-backend test -a)
# Give relayer account token balance
$main_cmd add-genesis-account ${RLY_ADDRESS_1} 500000000000ustrd

$main_cmd collect-gentxs 2> /dev/null
# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!STRIDE_DOCKER_NAMES[@]}; do
    node_name=${STRIDE_DOCKER_NAMES[i]}
    peers=""
    for j in "${!STRIDE_NODES[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${STRIDE_NODES[j]},${peers}"
        fi
    done
    sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$peers\"|g" "${STATE}/${node_name}/config/config.toml"
    # use blind address (not loopback) to allow incoming connections from outside networks for local debugging
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${node_name}/config/config.toml"
done

# make sure all Stride chains have the same genesis
for i in "${!STRIDE_CHAINS[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${main_node}/config/genesis.json ${STATE}/${STRIDE_DOCKER_NAMES[i]}/config/genesis.json
    fi
done
# Init Gaia
#############################################################################################################################
sh ${SCRIPT_DIR}/init_gaia.sh

# Spin up docker containers
#############################################################################################################################
docker-compose down
docker-compose up -d stride1 stride2 stride3 gaia1 gaia2 gaia3
echo "Chains creating..."
CSLEEP 10
echo "Restoring keys"
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_1" $main_chain
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_2" $main_gaia_chain

echo "creating hermes identifiers"
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw create-client $main_chain $main_gaia_chain > /dev/null
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw conn-init $main_chain $main_gaia_chain 07-tendermint-0 07-tendermint-0 > /dev/null

echo "Creating connection $main_chain <> $main_gaia_chain"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create connection $main_chain $main_gaia_chain > /dev/null

echo "Creating transfer channel"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer $main_gaia_chain connection-0 > /dev/null
# docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init $main_chain $main_gaia_chain connection-0 transfer transfer > /dev/null

echo "Starting hermes relayer"
docker-compose up --force-recreate -d hermes
echo "Waiting for hermes to be ready..."
CSLEEP 180

echo "Registering host zones..."
# first register host zone for ATOM chain
ATOM='uatom'
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
docker-compose --ansi never exec -T $main_node strided tx stakeibc register-host-zone connection-0 $ATOM $IBCATOM channel-0 --chain-id $main_chain --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
CSLEEP 180
echo "Registered host zones:"
docker-compose --ansi never exec -T $main_node strided q stakeibc list-host-zone

echo "\nBuild interchainquery relayer service (this takes ~120s...)"
rm -rf ./icq/keys
sh ${SCRIPT_DIR}/init_icq.sh

# sh ${SCRIPT_DIR}/tests/run_all_tests.sh
sh ${SCRIPT_DIR}/logs/create_logs.sh