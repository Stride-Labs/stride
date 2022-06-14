SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STATE=$SCRIPT_DIR/state
STRIDE_CHAIN=droplet
STRIDE_DOCKER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3 strideTestSeed)
STRIDE_PEER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3)
STRIDE_MONIKERS=(STRIDE_1 STRIDE_2 STRIDE_3 STRIDE_SEED)
STRIDE_ENDPOINTS=(stride_1.droplet.stridelabs.co stride_2.droplet.stridelabs.co stride_3.droplet.stridelabs.co seed.droplet.stridelabs.co)

VAL_TOKENS=500000000ustrd
STAKE_TOKENS=300000000ustrd

VAL_ACCTS=(val1 val2 val3 valseed)
MAIN_ID=0
main_node=${STRIDE_DOCKER_NAMES[$MAIN_ID]}
PORT_ID=26656

BASE_RUN=strided

ST_CMDS=()
for state_name in "${STRIDE_DOCKER_NAMES[@]}"; do
  ST_CMDS+=( "$BASE_RUN --home $STATE/$state_name" )
done
main_cmd=${ST_CMDS[$MAIN_ID]}

#   zone         = "us-central1-b"
# us-west1-a 
# us-east4-b
# us-central1-a
# e2-standard-4