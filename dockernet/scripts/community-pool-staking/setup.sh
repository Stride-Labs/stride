#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

path="dydx-noble"

relayer_logs=${LOGS}/relayer-${path}.log
relayer_config=$STATE/relayer-${path}/config
relayer_exec="$DOCKER_COMPOSE run --rm relayer-dydx-noble"

mkdir -p $relayer_config
chmod -R 777 $STATE/relayer-${path}
cp ${DOCKERNET_HOME}/config/relayer_config_dydx_noble.yaml $relayer_config/config.yaml

printf "DYDX <> NOBLE - Adding relayer keys..."
$relayer_exec rly keys restore dydx $RELAYER_DYDX_NOBLE_ACCT "$RELAYER_DYDX_NOBLE_MNEMONIC" >> $relayer_logs 2>&1
$relayer_exec rly keys restore noble $RELAYER_NOBLE_ACCT "$RELAYER_NOBLE_MNEMONIC" >> $relayer_logs 2>&1
echo "Done"

printf "DYDX <> NOBLE - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
$relayer_exec rly transact link ${path} >> $relayer_logs 2>&1
echo "Done"

$DOCKER_COMPOSE up -d relayer-${path}
SAVE_DOCKER_LOGS relayer-${path} $relayer_logs