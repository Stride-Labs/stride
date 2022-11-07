#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

HOST_CHAINS=(GAIA)

# cleanup any stale state
make stop-docker
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp
mkdir -p $SCRIPT_DIR/logs

HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log

# If we're testing an upgrade, setup cosmovisor
if [[ "$UPGRADE_NAME" != "" ]]; then
    printf "\n>>> UPGRADE ENABLED! ($UPGRADE_NAME)\n\n"
    
    # Update binary #2 with the binary that was just compiled
    mkdir -p $SCRIPT_DIR/upgrades/binaries
    rm -f $SCRIPT_DIR/upgrades/binaries/strided2
    cp $SCRIPT_DIR/../build/strided $SCRIPT_DIR/upgrades/binaries/strided2

    # Build a cosmovisor image with the old binary and replace the stride docker image with a new one
    #  that has both binaries and is running cosmovisor
    # The reason for having a separate cosmovisor image is so we can cache the building of cosmovisor and the old binary
    echo "Building Cosmovisor..."
    docker build \
        -t stridezone:cosmovisor \
        --build-arg old_commit_hash=$UPGRADE_OLD_COMMIT_HASH \
        -f ${SCRIPT_DIR}/upgrades/Dockerfile.cosmovisor .

    echo "Re-Building Stride with Upgrade Support..."
    docker build \
        -t stridezone:stride \
        --build-arg upgrade_name=$UPGRADE_NAME \
        -f ${SCRIPT_DIR}/upgrades/Dockerfile.stride .

    echo "Done"
fi

# Initialize the state for each chain
for chain_id in STRIDE ${HOST_CHAINS[@]}; do
    sh ${SCRIPT_DIR}/init_chain.sh $chain_id
done

# Start the chain and create the transfer channels
sh ${SCRIPT_DIR}/start_chain.sh STRIDE ${HOST_CHAINS[@]}
sh ${SCRIPT_DIR}/init_relayers.sh STRIDE ${HOST_CHAINS[@]}
sh ${SCRIPT_DIR}/create_channels.sh ${HOST_CHAINS[@]}

echo "Starting relayers"
docker-compose up -d hermes 
docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $HERMES_LOGS 2>&1 &

# Wait for hermes to start
( tail -f -n0 $HERMES_LOGS & ) | grep -q -E "Hermes has started"

# Register all host zones in parallel
pids=()
for i in ${!HOST_CHAINS[@]}; do
    if [[ "$i" != "0" ]]; then sleep 20; fi
    bash $SCRIPT_DIR/register_host.sh ${HOST_CHAINS[$i]} $i &
    pids[${i}]=$!
done
for i in ${!pids[@]}; do
    wait ${pids[$i]}
    echo "${HOST_CHAINS[$i]} - Done"
done

echo "Starting go relayers..."
for chain_id in ${HOST_CHAINS[@]}; do
    chain_name=$(printf "$chain_id" | awk '{ print tolower($0) }')

    docker-compose up -d relayer-${chain_name}
    docker-compose logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> ${LOGS}/relayer-${chain_name}.log 2>&1 &
done

$SCRIPT_DIR/create_logs.sh ${HOST_CHAINS[@]} &
