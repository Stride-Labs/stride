#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE ./icq/keys
docker-compose down

# build docker images
while getopts sghi flag; do
    case "${flag}" in
        s) docker build --tag stridezone:stride -f Dockerfile.stride . ;;
        g) docker build --tag stridezone:gaia -f Dockerfile.gaia . ;;
        h) docker build --tag stridezone:hermes -f Dockerfile.hermes . ;;
        i) docker build --tag stridezone:interchain-queries -f Dockerfile.icq . ;;
    esac
done

# TODO(TEST-117) Modularize/generalize chain init scripts 
# Initialize the state for stride/gaia and relayers
go build -mod=readonly -trimpath -o ~/go/bin ./..
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
sh ${SCRIPT_DIR}/init_hermes.sh
sh ${SCRIPT_DIR}/init_icq.sh

# Register host zone
# ICA staking test
# first register host zone for ATOM chain
# TODO(TEST-118) Improve integration test timing
echo "Registering host zones..."
ATOM='uatom'
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
CSLEEP 60
docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided tx stakeibc register-host-zone connection-0 $ATOM $IBCATOM "cosmos" channel-0 --chain-id $STRIDE_CHAIN --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
CSLEEP 120
echo "Registered host zones:"
docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided q stakeibc list-host-zone
sh ${SCRIPT_DIR}/tests/run_all_tests.sh
