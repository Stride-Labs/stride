#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
docker-compose down
rm -rf $SCRIPT_DIR/state $SCRIPT_DIR/logs/*.log $SCRIPT_DIR/logs/temp
mkdir -p $SCRIPT_DIR/logs

HERMES_LOGS=$SCRIPT_DIR/logs/hermes.log
ICQ_LOGS=$SCRIPT_DIR/logs/icq.log

# If we're testing an upgrade, setup cosmovisor
if [[ "$UPGRADE_NAME" != "" ]]; then
    mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/
    mkdir -p $SCRIPT_DIR/upgrades/cosmovisor/upgrades/$UPGRADE_NAME/bin/
    mkdir -p $SCRIPT_DIR/state/stride/cosmovisor

    rm -f $SCRIPT_DIR/upgrades/binaries/strided2
    cp $SCRIPT_DIR/../build/strided $SCRIPT_DIR/upgrades/binaries/strided2
    cp $SCRIPT_DIR/upgrades/binaries/strided1 $SCRIPT_DIR/upgrades/cosmovisor/genesis/bin/strided
    cp $SCRIPT_DIR/upgrades/binaries/strided2 $SCRIPT_DIR/upgrades/cosmovisor/upgrades/$UPGRADE_NAME/bin/strided

    # Build a cosmovisor image with the old binary and replace the stride docker image with a new one
    #  that has both binaries and is running cosmovisor
    docker build -t stridezone:cosmovisor --build-arg old_commit_hash=$UPGRADE_OLD_COMMIT_HASH -f ${SCRIPT_DIR}/upgrades/Dockerfile.cosmovisor .
    docker build -t stridezone:stride -f ${SCRIPT_DIR}/upgrades/Dockerfile.stride .
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