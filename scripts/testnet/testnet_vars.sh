SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "$1" ]
then
    echo "Error, you must pass testnet name in. E.g. \"sh setup_testnet_state.sh droplet\""
    exit 1
fi

echo "Setting up chain $1"

STRIDE_CHAIN=$1
STATE=$SCRIPT_DIR/$STRIDE_CHAIN/state
STRIDE_DOCKER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3 strideTestSeed)
STRIDE_PEER_NAMES=(strideTestNode1 strideTestNode2 strideTestNode3)
STRIDE_MONIKERS=(STRIDE_1 STRIDE_2 STRIDE_3 STRIDE_SEED)
STRIDE_ENDPOINTS=(stride_1.$STRIDE_CHAIN.stridelabs.co stride_2.$STRIDE_CHAIN.stridelabs.co stride_3.$STRIDE_CHAIN.stridelabs.co seed.$STRIDE_CHAIN.stridelabs.co)

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

GAIA_FOLDER="${STATE}/gaia"
GAIA_CMD="docker run -v ${STATE}/gaia:/gaia/.gaiad gcr.io/stride-nodes/testnet:tub_gaia gaiad --home /gaia/.gaiad"

GAIA_TOKENS=500000000uatom
GAIA_STAKE_TOKENS=300000000uatom
GAIA_ENDPOINT=gaia.$STRIDE_CHAIN.stridelabs.co
GAIA_CHAIN="GAIA_${STRIDE_CHAIN}"

HERMES_CMD="docker run -v ${STATE}/hermes.toml:/tmp/hermes.toml -v ${STATE}:/hermes/.hermes/keys gcr.io/stride-nodes/testnet:tub_hermes hermes -c /tmp/hermes.toml"
ICQ_CMD="docker run -v ${STATE}/icq:/hermes/.hermes/keys gcr.io/stride-nodes/testnet:tub_icq icq"

GETKEY() {
  grep -i -A 10 "\- name: $1" "$STATE/keys.txt" | tail -n 1
}

GETRLY2() {
  cat internal/state/keys.txt | tail -1
}

ICQ_STRIDE_KEY="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_KEY="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"

ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"