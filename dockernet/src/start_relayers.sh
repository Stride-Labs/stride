#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh


for chain in ${HOST_CHAINS[@]}; do
    chain_id=$(GET_VAR_VALUE     ${chain}_CHAIN_ID)
    relayer_exec=$(GET_VAR_VALUE RELAYER_${chain}_EXEC)
    chain_name=$(printf "$chain" | awk '{ print tolower($0) }')
    account_name=$(GET_VAR_VALUE RELAYER_${chain}_ACCT)
    mnemonic=$(GET_VAR_VALUE     RELAYER_${chain}_MNEMONIC)

    relayer_logs=${LOGS}/relayer-${chain_name}.log
    relayer_config=$STATE/relayer-${chain_name}/config

    mkdir -p $relayer_config
    chmod -R 777 $STATE/relayer-${chain_name}
    cp ${DOCKERNET_HOME}/config/relayer_config_stride.yaml $relayer_config/config.yaml

    printf "STRIDE <> $chain - Adding relayer keys..."
    $relayer_exec rly keys restore stride $RELAYER_STRIDE_ACCT "$mnemonic" >> $relayer_logs 2>&1
    $relayer_exec rly keys restore $chain_name $account_name "$mnemonic" >> $relayer_logs 2>&1
    echo "Done"

    printf "STRIDE <> $chain - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link stride-${chain_name} >> $relayer_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d relayer-${chain_name}
    $DOCKER_COMPOSE logs -f relayer-${chain_name} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
done

# If both dydx and noble are enabled, start their relayer
enable_dydx_noble_relayer=false
for chain1 in ${HOST_CHAINS[@]}; do 
    for chain2 in ${ACCESSORY_CHAINS[@]}; do 
        if [[ "$chain1" == "DYDX" && "$chain2" == "NOBLE" ]]; then
            enable_dydx_noble_relayer=true
        fi
    done
done
if [[ "$enable_dydx_noble_relayer" == "true" ]]; then
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
    $DOCKER_COMPOSE logs -f relayer-${path} | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> $relayer_logs 2>&1 &
fi
