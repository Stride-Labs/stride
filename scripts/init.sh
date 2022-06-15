# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for node_name in ${STRIDE_NODE_NAMES[@]}; do
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
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_1" $STRIDE_CHAIN
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_2" $GAIA_CHAIN

echo "creating hermes identifiers"
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw create-client $STRIDE_CHAIN $GAIA_CHAIN > /dev/null
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw conn-init $STRIDE_CHAIN $GAIA_CHAIN 07-tendermint-0 07-tendermint-0 > /dev/null

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create connection $STRIDE_CHAIN $GAIA_CHAIN > /dev/null

echo "Creating transfer channel"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 > /dev/null
# docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init $STRIDE_CHAIN $GAIA_CHAIN connection-0 transfer transfer > /dev/null

echo "Starting hermes relayer"
docker-compose up --force-recreate -d hermes
echo "Waiting for hermes to be ready..."

echo "\nBuild interchainquery relayer service (this takes ~120s...)"
rm -rf ./icq/keys
docker-compose build icq --no-cache
ICQ_RUN="docker-compose --ansi never run -T icq interchain-queries"

echo "\nAdd ICQ relayer addresses for Stride and Gaia:"
# TODO(TEST-82) redefine stride-testnet in lens' config to $STRIDE_CHAIN and gaia-testnet to $main-gaia-chain, then replace those below with $STRIDE_CHAIN and $GAIA_CHAIN
$ICQ_RUN keys restore test "$ICQ_STRIDE_KEY" --chain stride-testnet
$ICQ_RUN keys restore test "$ICQ_GAIA_KEY" --chain gaia-testnet

echo "\nICQ addresses for Stride and Gaia:"
# TODO(TEST-83) pull these addresses dynamically using jq
ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
# echo $ICQ_ADDRESS_STRIDE
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
# echo $ICQ_ADDRESS_GAIA

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
$STR1_EXEC tx bank send val1 $ICQ_ADDRESS_STRIDE 5000000ustrd --chain-id $STRIDE_CHAIN -y --keyring-backend test --home /stride/.strided
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
$GAIA1_EXEC tx bank send gval1 $ICQ_ADDRESS_GAIA 5000000uatom --chain-id $GAIA_CHAIN -y --keyring-backend test --home /gaia/.gaiad

echo "\nLaunch interchainquery relayer"
docker-compose up --force-recreate -d icq

# Register host zone
# ICA staking test
# first register host zone for ATOM chain
ATOM='uatom'
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
CSLEEP 10
docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided tx stakeibc register-host-zone connection-0 $ATOM $IBCATOM channel-0 --chain-id $STRIDE_CHAIN --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
CSLEEP 30
sh ${SCRIPT_DIR}/tests/run_all_tests.sh
