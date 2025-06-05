#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

for chain in ${HOST_CHAINS[@]:-}; do
    chain_id=$(GET_VAR_VALUE     ${chain}_CHAIN_ID)
    chain_name=$(printf "$chain" | awk '{ print tolower($0) }')
    account_name=$(GET_VAR_VALUE RELAYER_${chain}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain}_MNEMONIC)

    hermes_logs=${LOGS}/hermes.log
    hermes_config=$STATE/hermes
    hermes_exec="$DOCKER_COMPOSE run --rm hermes hermes"

    mkdir -p $hermes_config
    cp ${DOCKERNET_HOME}/config/hermes_config.toml $hermes_config/config.toml
    echo "$mnemonic" > $hermes_config/mnemonic.txt
    chmod -R 777 $hermes_config

    printf "STRIDE <> $chain - Adding hermes keys..."
    $hermes_exec keys add --chain STRIDE --mnemonic-file /home/hermes/.hermes/mnemonic.txt >> $hermes_logs 2>&1
    $hermes_exec keys add --chain $chain_id --mnemonic-file /home/hermes/.hermes/mnemonic.txt >> $hermes_logs 2>&1
    echo "Done"

    printf "STRIDE <> $chain - Creating client, connection, and transfer channel..." | tee -a $hermes_logs
    $hermes_exec create channel --a-chain STRIDE --b-chain $chain_id \
            --a-port transfer --b-port transfer --new-client-connection --yes >> $hermes_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d hermes
    SAVE_DOCKER_LOGS hermes $hermes_logs
done