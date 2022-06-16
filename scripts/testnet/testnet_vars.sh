SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "$1" ]
then
    echo "Error, you must pass testnet name in. E.g. \"sh setup_testnet_state.sh droplet\""
    exit 1
fi

source $SCRIPT_DIR/$1/vars.sh

if [ -z "${STRIDE_CHAIN}" ]
then
    echo "STRIDE_CHAIN not defined. Did you pass in your desired testnet? E.g. \"sh setup_testnet_state.sh droplet\""
    exit 1
fi

echo "Setting up chain $STRIDE_CHAIN"

STRIDE_CHAIN=droplet
STATE=$SCRIPT_DIR/$STRIDE_CHAIN/state
STRIDE_DOCKER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3 strideTestSeed)
STRIDE_PEER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3)
STRIDE_MONIKERS=(STRIDE_1 STRIDE_2 STRIDE_3 STRIDE_SEED)
STRIDE_ENDPOINTS=(stride_1.droplet.stridelabs.co stride_2.droplet.stridelabs.co stride_3.droplet.stridelabs.co seed.droplet.stridelabs.co)

VAL_TOKENS=500000000ustrd
STAKE_TOKENS=300000000ustrd

VAL_ACCTS=(val1 val2 val3 valseed)
SEED_ID=3
MAIN_ID=0
main_node=${STRIDE_DOCKER_NAMES[$MAIN_ID]}
seed_node=${STRIDE_DOCKER_NAMES[$SEED_ID]}
PORT_ID=26656

BASE_RUN=strided

ST_CMDS=()
for state_name in "${STRIDE_DOCKER_NAMES[@]}"; do
  ST_CMDS+=( "$BASE_RUN --home $STATE/$state_name" )
done
main_cmd=${ST_CMDS[$MAIN_ID]}
