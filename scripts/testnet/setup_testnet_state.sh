SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh $1

# run through init args, if needed
while getopts b flag; do
    case "${flag}" in
        b) ignite chain init ;;
    esac
done

echo "Cleaning state"
rm -rf $STATE
mkdir $STATE
touch $STATE/keys.txt

# Init Stride
#############################################################################################################################
# fetch the stride node ids
STRIDE_NODES=()
SEED_NODE_ID=""
# then, we initialize our chains 
echo 'Initializing chains...'
for i in ${!STRIDE_DOCKER_NAMES[@]}; do
    node_name=${STRIDE_DOCKER_NAMES[i]}
    moniker=${STRIDE_MONIKERS[i]}
    val_acct=${VAL_ACCTS[i]}
    st_cmd=${ST_CMDS[i]}
    $st_cmd init $moniker --chain-id $STRIDE_CHAIN --overwrite 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${node_name}/config/genesis.json"
    # now we grab the relevant node id
    end_name=${STRIDE_ENDPOINTS[i]}
    node_id=$($st_cmd tendermint show-node-id)@$end_name:$PORT_ID
    if [ $i -ne $SEED_ID ]
    then
        # add VALidator account
        $st_cmd keys add $val_acct --keyring-backend=test >> $STATE/keys.txt 2>&1
        # get validator address
        VAL_ADDR=$($st_cmd keys show $val_acct --keyring-backend test -a)
        # add money for this validator account
        $st_cmd add-genesis-account ${VAL_ADDR} $VAL_TOKENS
        # actually set this account as a validator
        yes | $st_cmd gentx $val_acct $STAKE_TOKENS --chain-id $STRIDE_CHAIN --keyring-backend test 2> /dev/null
    else
        SEED_NODE_ID=$node_id
    fi
   
    echo $node_id
    echo "$node_name id: $node_id" >> $STATE/keys.txt
    STRIDE_NODES+=( $node_id )
    if [ $i -ne $MAIN_ID ] && [ $i -ne $SEED_ID ]
    then
        $main_cmd add-genesis-account ${VAL_ADDR} $VAL_TOKENS
        cp ${STATE}/${node_name}/config/gentx/*.json ${STATE}/${main_node}/config/gentx/
    fi
done

# # now we process gentx txs 
$main_cmd collect-gentxs 2> /dev/null

# set seed node to actually be a seed node
sed -i -E 's|seed_mode = false|seed_mode = true|g' "${STATE}/${seed_node}/config/config.toml"


# add peers in config.toml so that nodes can find each other by constructing a fully connected
# graph of nodes
for i in ${!STRIDE_DOCKER_NAMES[@]}; do
    node_name=${STRIDE_DOCKER_NAMES[i]}
    peers=""
    for j in "${!STRIDE_DOCKER_NAMES[@]}"; do
        if [ $j -ne $i ] && [ ${STRIDE_DOCKER_NAMES[j]} != strideTestSeed ]
        then
            peers="${STRIDE_NODES[j]},${peers}"
        fi
    done
    sed -i -E "s|seeds = .*|seeds = \"$SEED_NODE_ID\"|g" "${STATE}/${node_name}/config/config.toml"
    sed -i -E "s|persistent_peers = .*|persistent_peers = \"$peers\"|g" "${STATE}/${node_name}/config/config.toml"
    sed -i -E "s|cors_allowed_origins = \[\]|cors_allowed_origins = [\"\*\"]|g" "${STATE}/${node_name}/config/config.toml"
    sed -i -E "s|127.0.0.1|0.0.0.0|g" "${STATE}/${node_name}/config/config.toml"
done

# make sure all Stride chains have the same genesis
for i in "${!STRIDE_DOCKER_NAMES[@]}"; do
    if [ $i -ne $MAIN_ID ]
    then
        cp ${STATE}/${main_node}/config/genesis.json ${STATE}/${STRIDE_DOCKER_NAMES[i]}/config/genesis.json
    fi
done
