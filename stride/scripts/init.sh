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
    mkdir -p $STATE/$chain_name
done

# run through init args, if needed
while getopts bdfsa flag; do
    case "${flag}" in
        b) ignite chain build ;;
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
    vkey=${VKEYS[i]}
    val_acct=${VAL_ACCTS[i]}
    st_cmd=${ST_CMDS[i]}
    echo "\t$chain_name"
    $st_cmd init test --chain-id $chain_name --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${chain_name}/config/genesis.json"
    # add VALidator account
    echo $vkey | $st_cmd keys add $val_acct --recover --keyring-backend=test > /dev/null
    # get validator address
    VAL_ADDR=$($st_cmd keys show $val_acct --keyring-backend test -a)
    # add money for this validator account
    $st_cmd add-genesis-account ${VAL_ADDR} 500000000000ustrd
    # actually set this account as a validator
    yes | $st_cmd gentx $val_acct 1000000000ustrd --chain-id $main_chain --keyring-backend test
    # now we process these txs 
    $st_cmd collect-gentxs 2> /dev/null
    # now we grab the relevant node id
    dock_name=${STRIDE_DOCKER_NAMES[i]}
    node_id=$($st_cmd tendermint show-node-id)@$dock_name:$PORT_ID
    STRIDE_NODES+=( $node_id )

    if [ $i -ne $MAIN_ID ]
    then
        $main_cmd add-genesis-account ${VAL_ADDR} 500000000000ustrd
        cp ${STATE}/${chain_name}/config/gentx/*.json ${STATE}/${main_chain}/config/gentx/
    fi
done

# Restore relayer account on stride
echo $RLY_MNEMONIC_1 | $main_cmd keys add rly1 --recover --keyring-backend=test > /dev/null
RLY_ADDRESS_1=$($main_cmd keys show rly1 --keyring-backend test -a)
# Give relayer account token balance
$main_cmd add-genesis-account ${RLY_ADDRESS_1} 500000000000ustrd

$main_cmd collect-gentxs 2> /dev/null
# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!STRIDE_CHAINS[@]}; do
    chain_name=${STRIDE_CHAINS[i]}
    peers=""
    for j in "${!STRIDE_NODES[@]}"; do
        if [ $j -ne $i ]
        then
            peers="${STRIDE_NODES[j]},${peers}"
        fi
    done
    echo "${chain_name} peers are:"
    echo $peers
    sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$peers\"|g" "${STATE}/${chain_name}/config/config.toml"
    # use blind address (not loopback) to allow incoming connections from outside networks for local debugging
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${chain_name}/config/config.toml"
done

# make sure all Stride chains have the same genesis
for i in "${!STRIDE_CHAINS[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${main_chain}/config/genesis.json ${STATE}/${STRIDE_CHAINS[i]}/config/genesis.json
    fi
done

# Init Gaia
#############################################################################################################################
sh ${SCRIPT_DIR}/init_gaia.sh

# Accounts and exec commands
#############################################################################################################################

# Stride
echo '\nStride addresses 1,2,3'
STRIDE_ADDRESS_1=$($BASE_RUN keys show val1 --home $STATE/STRIDE_1 --keyring-backend test -a)
STRIDE_ADDRESS_2=$($BASE_RUN keys show val2 --home $STATE/STRIDE_2 --keyring-backend test -a)
STRIDE_ADDRESS_3=$($BASE_RUN keys show val3 --home $STATE/STRIDE_3 --keyring-backend test -a)
echo "\t${STRIDE_ADDRESS_1}"
echo "\t${STRIDE_ADDRESS_2}"
echo "\t${STRIDE_ADDRESS_3}"

# Gaia
echo '\nGaia addresses 1,2,3'
GAIA_ADDRESS_1=$($GAIA_RUN gaia1 gaiad keys show gval1 --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_2=$($GAIA_RUN gaia2 gaiad keys show gval2 --keyring-backend test -a --home=/gaia/.gaiad)
GAIA_ADDRESS_3=$($GAIA_RUN gaia3 gaiad keys show gval3 --keyring-backend test -a --home=/gaia/.gaiad)
echo "\t${GAIA_ADDRESS_1}"
echo "\t${GAIA_ADDRESS_2}"
echo "\t${GAIA_ADDRESS_3}"

# Relayers
echo '\nRelay addresses 1,2'
# NOTE: using $main_cmd and $main_gaia_cmd here ONLY works because they rly1 and rly2
# keys are on stride1 and gaia1, respectively
RLY_ADDRESS_1=$($main_cmd keys show rly1 --keyring-backend test -a)
RLY_ADDRESS_2=$($main_gaia_cmd keys show rly2 --keyring-backend test -a)
echo "\t$RLY_ADDRESS_1"
echo "\t$RLY_ADDRESS_2"

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided"
STR2_EXEC="docker-compose --ansi never exec -T stride2 strided"
STR3_EXEC="docker-compose --ansi never exec -T stride3 strided"

# Spin up docker containers
#############################################################################################################################
# strided start --home state/STRIDE_1  # TESTING ONLY
sleep 5
docker-compose down
docker-compose up -d stride1 stride2 stride3 gaia1 gaia2 gaia3
echo "Chains created"
sleep 10
echo "Restoring keys"
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_1" $main_chain
docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_2" $main_gaia_chain
sleep 10
echo "Creating transfer channel"

echo "creating hermes identifiers"
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw create-client $main_chain $main_gaia_chain > /dev/null
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw conn-init $main_chain $main_gaia_chain 07-tendermint-0 07-tendermint-0 > /dev/null
docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init $main_chain $main_gaia_chain connection-0 transfer transfer > /dev/null

echo "Creating connection $main_chain <> $main_gaia_chain"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create connection $main_chain $main_gaia_chain > /dev/null
echo "Connection created"
echo "Creating transfer channel"
docker-compose run -T hermes hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer $main_gaia_chain connection-0 > /dev/null
echo "Tranfer channel created"

docker-compose down hermes
echo "Starting hermes relayer"
docker-compose up --force-recreate -d hermes

# IBC token transfer tests
#############################################################################################################################
source ${SCRIPT_DIR}/ibc_token_transfer.sh
