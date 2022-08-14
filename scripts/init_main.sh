#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

BUILDDIR=$SCRIPT_DIR/../build

# check and install dependencies
DEPENDENCIES="jq bats"
deps=0
for name in $DEPENDENCIES; do
    [[ $(type $name 2>/dev/null) ]] || { echo "\n    * $name is required to run this script;";deps=1; }
done
if [[ "$deps" == "1" ]]; then 
    echo "Install the missing dependencies and rerun this script..."
    exit 1;
fi

# cleanup any stale state
rm -rf $STATE 
docker-compose down

# build docker images and local binaries
while getopts sghi flag; do
    case "${flag}" in
        s) printf '%s' "Building Stride Docker...  ";
           docker build --tag stridezone:stride -f Dockerfile.stride . ;

           printf '%s' "Building Stride Locally...  ";
           go build -mod=readonly -trimpath -o $BUILDDIR ./... ;
           echo "Done" ;;

        g) printf '%s' "Building Gaia Docker...    ";
           docker build --tag stridezone:gaia -f Dockerfile.gaia . ;

           printf '%s' "Building Gaia Locally...   ";
           cd deps/gaia ; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true ;
           cd ../.. ;
           echo "Done" ;;

        h) printf '%s' "Building Hermes Docker... ";
           docker build --tag stridezone:hermes -f Dockerfile.hermes . ;

           printf '%s' "Building Hermes Locally... ";
           cd deps/hermes; 
           cargo build --release --target-dir $BUILDDIR/hermes; 
           cd ../..
           echo "Done" ;;

        i) printf '%s' "Building ICQ Docker...    ";
           docker build --tag stridezone:interchain-queries -f Dockerfile.icq . ;

           printf '%s' "Building ICQ Locally...    ";
           cd deps/interchain-queries; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
           echo "Done" ;;           
    esac
done

# TODO(TEST-117) Modularize/generalize chain init scripts 
# Initialize the state for stride/gaia and relayers
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
sh ${SCRIPT_DIR}/init_relayers.sh

echo "Creating stride chain"
docker-compose up -d stride1 stride2 stride3 

echo "Creating gaia chain"
docker-compose up -d gaia1 gaia2 gaia3

# Register host zone
# ICA staking test
# first register host zone for ATOM chain
# TODO(TEST-118) Improve integration test timing
# echo "Registering host zones..."
# ATOM='uatom'
# IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
# CSLEEP 60
# docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided tx stakeibc register-host-zone connection-0 $ATOM $IBCATOM "cosmos" channel-0 --chain-id $STRIDE_CHAIN --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
# CSLEEP 120
# echo "Registered host zones:"
# docker-compose --ansi never exec -T $STRIDE_MAIN_NODE strided q stakeibc list-host-zone
# sh ${SCRIPT_DIR}/tests/run_all_tests.sh