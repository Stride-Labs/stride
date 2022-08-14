#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# cleanup any stale state
rm -rf $STATE 
docker-compose down

# Initialize the state for stride/gaia and relayers
sh ${SCRIPT_DIR}/init_stride.sh
sh ${SCRIPT_DIR}/init_gaia.sh
sh ${SCRIPT_DIR}/init_relayers.sh

echo "Starting STRIDE chain"
docker-compose up -d stride1 stride2 stride3 

echo "Starting GAIA chain"
docker-compose up -d gaia1 gaia2 gaia3

echo "Starting relayers"
docker-compose up -d hermes icq

printf "Waiting for STRIDE and GAIA to start..."
( docker-compose logs -f stride1 & ) | grep -q "finalizing commit of block"
( docker-compose logs -f gaia1 & ) | grep -q "finalizing commit of block"
sleep 5
echo "Done"

printf "Creating connection..."
$HERMES_EXEC create connection $STRIDE_CHAIN_ID $GAIA_CHAIN_ID &> /dev/null
echo "Done"

printf "Creating transfer channel..."
$HERMES_EXEC create channel --port-a transfer --port-b transfer $GAIA_CHAIN_ID connection-0 &> /dev/null
echo "Done"

bash $SCRIPT_DIR/register_host.sh